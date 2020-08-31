package cmd

import (
	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/spf13/cobra"
)

var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Test commands that directly do things with the launchpayload",
	Long:  ``,
}

var payloadTestCmd = &cobra.Command{
	Use:   "payloadtest",
	Short: "Test commands that directly do things with the launchpayload",
	Long:  ``,
	Run:   initCLI,
}

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(payloadTestCmd)
}

func initDaemon(cmd *cobra.Command, args []string) {
	lctrld.InitDaemon(settings, args[0])
}

func initCLI(cmd *cobra.Command, args []string) {
	lctrld.GenerateKeys(settings, args[0])
}
