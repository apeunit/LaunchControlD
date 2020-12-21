package cmd

import (
	"log"

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
	if err := server.ServeHTTP(settings); err != nil {
		log.Fatal(err)
	}
}
