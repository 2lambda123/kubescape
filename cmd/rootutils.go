package cmd

import (
	"fmt"
	"os"
	"strings"

	v1 "github.com/kubescape/backend/pkg/client/v1"
	logger "github.com/kubescape/go-logger"
	"github.com/kubescape/go-logger/helpers"
	"github.com/kubescape/kubescape/v2/core/cautils/getter"

	"github.com/mattn/go-isatty"
)

const envFlagUsage = "Send report results to specific URL. Format:<ReportReceiver>,<Backend>,<Frontend>.\n\t\tExample:report.armo.cloud,api.armo.cloud,portal.armo.cloud"

func initLogger() {
	logger.DisableColor(rootInfo.DisableColor)
	logger.EnableColor(rootInfo.EnableColor)

	if rootInfo.LoggerName == "" {
		if l := os.Getenv("KS_LOGGER_NAME"); l != "" {
			rootInfo.LoggerName = l
		} else {
			if isatty.IsTerminal(os.Stdout.Fd()) {
				rootInfo.LoggerName = "pretty"
			} else {
				rootInfo.LoggerName = "zap"
			}
		}
	}

	logger.InitLogger(rootInfo.LoggerName)

}
func initLoggerLevel() {
	if rootInfo.Logger == helpers.InfoLevel.String() {
	} else if l := os.Getenv("KS_LOGGER"); l != "" {
		rootInfo.Logger = l
	}

	if err := logger.L().SetLevel(rootInfo.Logger); err != nil {
		logger.L().Fatal(fmt.Sprintf("supported levels: %s", strings.Join(helpers.SupportedLevels(), "/")), helpers.Error(err))
	}
}

func initCacheDir() {
	if rootInfo.CacheDir != getter.DefaultLocalStore {
		getter.DefaultLocalStore = rootInfo.CacheDir
	} else if cacheDir := os.Getenv("KS_CACHE_DIR"); cacheDir != "" {
		getter.DefaultLocalStore = cacheDir
	} else {
		return // using default cache dir location
	}

	logger.L().Debug("cache dir updated", helpers.String("path", getter.DefaultLocalStore))
}
func initEnvironment() {
	if rootInfo.KSCloudBEURLs == "" {
		rootInfo.KSCloudBEURLs = rootInfo.KSCloudBEURLsDep
	}
	urlSlices := strings.Split(rootInfo.KSCloudBEURLs, ",")
	if len(urlSlices) != 1 && len(urlSlices) < 2 {
		logger.L().Fatal("expected at least 2 URLs (report, api)")
	}
	switch len(urlSlices) {
	case 1:
		switch urlSlices[0] {
		case "dev", "development":
			getter.SetKSCloudAPIConnector(getter.NewKSCloudAPIDev())
		case "stage", "staging":
			getter.SetKSCloudAPIConnector(getter.NewKSCloudAPIStaging())
		case "":
			getter.SetKSCloudAPIConnector(getter.NewKSCloudAPIProd())
		default:
			logger.L().Fatal("--environment flag usage: " + envFlagUsage)
		}
	case 2:
		ksEventReceiverURL := urlSlices[0] // mandatory
		ksBackendURL := urlSlices[1]       // mandatory

		getter.SetKSCloudAPIConnector(v1.NewKSCloudAPICustomized(
			ksBackendURL,
			v1.WithReportURL(ksEventReceiverURL),
		))
	}
}
