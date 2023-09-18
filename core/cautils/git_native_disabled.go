//go:build !gitenabled

package cautils

import (
	"fmt"
	"time"

	"github.com/kubescape/go-git-url/apis"
	"github.com/matthyx/go-gitlog"
)

type gitRepository struct {
	gitRepo          gitlog.GitLog
	fileToLastCommit map[string]*gitlog.Commit
}

func newGitRepository(root string) (*gitRepository, error) {
	gitRepo := gitlog.New(&gitlog.Config{Path: root})

	return &gitRepository{
		gitRepo: gitRepo,
	}, nil
}

func (g *gitRepository) GetFileLastCommit(filePath string) (*apis.Commit, error) {
	if len(g.fileToLastCommit) == 0 {
		filePathToCommitTime := map[string]time.Time{}
		filePathToCommit := map[string]*gitlog.Commit{}
		allCommits, _ := g.gitRepo.Log(nil, nil)

		// builds a map of all files to their last commit
		for _, commit := range allCommits {
			// Ignore merge commits (2+ parents)
			if len(commit.Parents) > 1 {
				continue
			}

			for _, file := range commit.Files {
				commitTime := commit.Author.Date

				// In case we have the commit information for the file which is not the latest - we override it
				if currentCommitTime, exists := filePathToCommitTime[file]; exists {
					if currentCommitTime.Before(commitTime) {
						filePathToCommitTime[file] = commitTime
						filePathToCommit[file] = commit
					}
				} else {
					filePathToCommitTime[file] = commitTime
					filePathToCommit[file] = commit
				}
			}
		}

		g.fileToLastCommit = filePathToCommit
	}

	if relevantCommit, exists := g.fileToLastCommit[filePath]; exists {
		return g.getCommit(relevantCommit), nil
	}

	return nil, fmt.Errorf("failed to get commit information for file: %s", filePath)
}

func (g *gitRepository) getCommit(commit *gitlog.Commit) *apis.Commit {
	return &apis.Commit{
		SHA: commit.Hash.Long,
		Author: apis.Committer{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			Date:  commit.Author.Date,
		},
		Message: commit.Subject + "\n" + commit.Body,
		Committer: apis.Committer{
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
			Date:  commit.Committer.Date,
		},
		Files: []apis.Files{},
	}
}
