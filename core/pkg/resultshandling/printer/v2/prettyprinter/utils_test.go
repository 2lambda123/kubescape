package prettyprinter

import (
	"reflect"
	"testing"

	"github.com/kubescape/kubescape/v2/core/pkg/resultshandling/printer/v2/prettyprinter/tableprinter/imageprinter"
	"github.com/kubescape/opa-utils/reporthandling/results/v1/reportsummary"
	"github.com/stretchr/testify/assert"
)

func Test_filterComplianceFrameworks(t *testing.T) {
	tests := []struct {
		name                   string
		summaryDetails         *reportsummary.SummaryDetails
		expectedSummaryDetails *reportsummary.SummaryDetails
	}{
		{
			name: "check compliance frameworks are filtered",
			summaryDetails: &reportsummary.SummaryDetails{
				Frameworks: []reportsummary.FrameworkSummary{
					{
						Name: "CIS Kubernetes Benchmark",
					},
					{
						Name: "nsa",
					},
					{
						Name: "mitre",
					},
				},
			},
			expectedSummaryDetails: &reportsummary.SummaryDetails{
				Frameworks: []reportsummary.FrameworkSummary{
					{
						Name: "nsa",
					},
					{
						Name: "mitre",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complianceFws := filterComplianceFrameworks(tt.summaryDetails.ListFrameworks())
			assert.True(t, reflect.DeepEqual(complianceFws, tt.expectedSummaryDetails.ListFrameworks()))
		})
	}
}

func Test_getWorkloadPrefixForCmd(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		kind      string
		name1     string
		want      string
	}{
		{
			name:      "non-empty namespace",
			namespace: "default",
			kind:      "pod",
			name1:     "test",
			want:      "namespace: default, name: test, kind: pod",
		},
		{
			name:      "empty namespace",
			namespace: "",
			kind:      "pod",
			name1:     "test",
			want:      "name: test, kind: pod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getWorkloadPrefixForCmd(tt.namespace, tt.kind, tt.name1); got != tt.want {
				t.Errorf("getWorkloadPrefixForCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTopWorkloadsTitle(t *testing.T) {
	title := getTopWorkloadsTitle(0)
	assert.Equal(t, "", title)

	title = getTopWorkloadsTitle(1)
	assert.Equal(t, "Your most risky workload:\n", title)

	title = getTopWorkloadsTitle(2)
	assert.Equal(t, "Your most risky workloads:\n", title)

	title = getTopWorkloadsTitle(10)
	assert.Equal(t, "Your most risky workloads:\n", title)
}

func Test_getSeverityToSummaryMap(t *testing.T) {
	tests := []struct {
		name           string
		summaryDetails imageprinter.ImageScanSummary
		expected       map[string]imageprinter.SeveritySummary
		shouldMerge    bool
	}{
		{
			name: "without merging",
			summaryDetails: imageprinter.ImageScanSummary{
				MapsSeverityToSummary: map[string]*imageprinter.SeveritySummary{
					"High": {
						NumberOfCVEs:        10,
						NumberOfFixableCVEs: 2,
					},
					"Low": {
						NumberOfCVEs:        5,
						NumberOfFixableCVEs: 1,
					},
					"Negligible": {
						NumberOfCVEs:        3,
						NumberOfFixableCVEs: 0,
					},
				},
			},
			shouldMerge: false,
			expected: map[string]imageprinter.SeveritySummary{
				"High": {
					NumberOfCVEs:        10,
					NumberOfFixableCVEs: 2,
				},
				"Low": {
					NumberOfCVEs:        5,
					NumberOfFixableCVEs: 1,
				},
				"Negligible": {
					NumberOfCVEs:        3,
					NumberOfFixableCVEs: 0,
				},
			},
		},
		{
			name: "with merging",
			summaryDetails: imageprinter.ImageScanSummary{
				MapsSeverityToSummary: map[string]*imageprinter.SeveritySummary{
					"Critical": {
						NumberOfCVEs:        15,
						NumberOfFixableCVEs: 2,
					},
					"High": {
						NumberOfCVEs:        10,
						NumberOfFixableCVEs: 2,
					},
					"Medium": {
						NumberOfCVEs:        5,
						NumberOfFixableCVEs: 1,
					},
					"Low": {
						NumberOfCVEs:        5,
						NumberOfFixableCVEs: 1,
					},
					"Negligible": {
						NumberOfCVEs:        3,
						NumberOfFixableCVEs: 0,
					},
				},
			},
			shouldMerge: true,
			expected: map[string]imageprinter.SeveritySummary{
				"Critical": {
					NumberOfCVEs:        15,
					NumberOfFixableCVEs: 2,
				},
				"High": {
					NumberOfCVEs:        10,
					NumberOfFixableCVEs: 2,
				},
				"Medium": {
					NumberOfCVEs:        5,
					NumberOfFixableCVEs: 1,
				},
				"Other": {
					NumberOfCVEs:        8,
					NumberOfFixableCVEs: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sevToSummaryMap := getSeverityToSummaryMap(tt.summaryDetails, tt.shouldMerge)

			for k, v := range sevToSummaryMap {
				if v.NumberOfCVEs != tt.expected[k].NumberOfCVEs || v.NumberOfFixableCVEs != tt.expected[k].NumberOfFixableCVEs {
					t.Errorf("in test: %v, error for key %v, want: %v, have :%v", tt.name, k, tt.expected[k], v)
				}
			}
		})
	}
}

func Test_filterCVEsBySeverities(t *testing.T) {
	test := []struct {
		name         string
		cves         []imageprinter.CVE
		severities   []string
		expectedCVEs []imageprinter.CVE
	}{
		{
			name: "empty severities list",
			cves: []imageprinter.CVE{
				{
					Severity: "High",
					ID:       "CVE-2020-1234",
				},
			},
			severities:   []string{},
			expectedCVEs: []imageprinter.CVE{},
		},
		{
			name: "one severity",
			cves: []imageprinter.CVE{
				{
					Severity: "High",
					ID:       "CVE-2020-1234",
				},
				{
					Severity: "Medium",
					ID:       "CVE-2020-1235",
				},
			},
			severities: []string{"High"},
			expectedCVEs: []imageprinter.CVE{
				{
					Severity: "High",
					ID:       "CVE-2020-1234",
				},
			},
		},
		{
			name: "multiple severities",
			cves: []imageprinter.CVE{
				{
					Severity: "High",
					ID:       "CVE-2020-1234",
				},
				{
					Severity: "Medium",
					ID:       "CVE-2020-1235",
				},
				{
					Severity: "Low",
					ID:       "CVE-2020-1236",
				},
			},
			severities: []string{"High", "Low"},
			expectedCVEs: []imageprinter.CVE{
				{
					Severity: "High",
					ID:       "CVE-2020-1234",
				},
				{
					Severity: "Low",
					ID:       "CVE-2020-1236",
				},
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			filteredCVEs := filterCVEsBySeverities(tt.cves, tt.severities)

			for i := range filteredCVEs {
				if filteredCVEs[i].Severity != tt.expectedCVEs[i].Severity || filteredCVEs[i].ID != tt.expectedCVEs[i].ID {
					t.Errorf("filterCVEsBySeverities() = %v, want %v", filteredCVEs, tt.expectedCVEs)
				}
			}

		})
	}

}

func Test_sortTopVulnerablePackages(t *testing.T) {
	tests := []struct {
		name              string
		pkgScores         map[string]*imageprinter.PackageScore
		expectedPkgScores map[string]*imageprinter.PackageScore
	}{
		{
			name: "change order",
			pkgScores: map[string]*imageprinter.PackageScore{
				"pkg1": {
					Version: "1.0.0",
					Score:   10,
				},
				"pkg2": {
					Version: "2.0.0",
					Score:   20,
				},
				"pkg3": {
					Version: "3.0.0",
					Score:   15,
				},
			},
			expectedPkgScores: map[string]*imageprinter.PackageScore{
				"pkg2": {
					Version: "2.0.0",
					Score:   20,
				},
				"pkg3": {
					Version: "3.0.0",
					Score:   15,
				},
				"pkg1": {
					Version: "1.0.0",
					Score:   10,
				},
			},
		},
		{
			name: "keep order",
			pkgScores: map[string]*imageprinter.PackageScore{
				"pkg1": {
					Version: "1.0.0",
					Score:   30,
				},
				"pkg2": {
					Version: "2.0.0",
					Score:   20,
				},
				"pkg3": {
					Version: "3.0.0",
					Score:   10,
				},
			},
			expectedPkgScores: map[string]*imageprinter.PackageScore{
				"pkg1": {
					Version: "1.0.0",
					Score:   30,
				},
				"pkg2": {
					Version: "2.0.0",
					Score:   20,
				},
				"pkg3": {
					Version: "3.0.0",
					Score:   10,
				},
			},
		},
		{
			name:              "empty",
			pkgScores:         map[string]*imageprinter.PackageScore{},
			expectedPkgScores: map[string]*imageprinter.PackageScore{},
		},
	}

	for _, tt := range tests {
		actual := sortTopVulnerablePackages(tt.pkgScores)
		for k, v := range actual {
			if v.Version != tt.expectedPkgScores[k].Version || v.Score != tt.expectedPkgScores[k].Score {
				t.Errorf("in test: %v, error for key %v, want: %v, have :%v", tt.name, k, tt.expectedPkgScores[k], v)
			}
		}
	}
}
