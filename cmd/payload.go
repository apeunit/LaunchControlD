package cmd

import (
	"fmt"

	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/spf13/cobra"
)

var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Test commands that directly do things with the launchpayload",
	Long:  ``,
}

var generateKeysCmd = &cobra.Command{
	Use:   "generatekeys",
	Short: "Test commands that directly do things with the launchpayload",
	Long:  ``,
	Run:   generateKeys,
}

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(generateKeysCmd)
}

func generateKeys(cmd *cobra.Command, args []string) {
	err := lctrld.InitDaemon(settings, args[0])
	fmt.Println("GenerateKeys err:", err)
}
