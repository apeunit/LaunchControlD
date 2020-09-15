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

var setupChainCmd = &cobra.Command{
	Use:   "setup",
	Short: "Does everything to setup a Cosmos-SDK based chain",
	Long:  ``,
	Run:   setupChain,
}

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(setupChainCmd)
}

func setupChain(cmd *cobra.Command, args []string) {
	lctrld.InitDaemon(settings, args[0])
	lctrld.GenerateKeys(settings, args[0])
	lctrld.AddGenesisAccounts(settings, args[0])
	lctrld.GenesisTxs(settings, args[0])
	lctrld.CollectGenesisTxs(settings, args[0])
	lctrld.EditConfigs(settings, args[0])
}
