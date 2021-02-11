package cmd

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/makasim/sentryhook"
	log "github.com/sirupsen/logrus"

	"github.com/apeunit/LaunchControlD/pkg/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serve the REST API for the project",
	Long:  ``,
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve(cmd *cobra.Command, args []string) {
	// enable sentry logging
	if settings.Sentry.Enabled {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:         settings.Sentry.DSN,
			Release:     fmt.Sprint("lctrld@", settings.RuntimeVersion),
			Environment: settings.Sentry.Environment,
		})
		if err != nil {
			log.Warnf("sentry configuration failed, error will not be reported: %v", err)
		}
		// add hook to sentry for logging events
		log.AddHook(sentryhook.New([]log.Level{
			log.PanicLevel,
			log.FatalLevel,
			log.ErrorLevel,
			log.WarnLevel}))
	} else {
		log.Info("log reporting via sentry is disabled")
	}

	if err := server.ServeHTTP(settings); err != nil {
		log.Error(err)
	}
}
