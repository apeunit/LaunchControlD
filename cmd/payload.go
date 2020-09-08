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

var initDaemonCmd = &cobra.Command{
	Use:   "initdaemon",
	Short: "Runs lctrld.InitDaemon",
	Long:  ``,
	Run:   initDaemon,
}

var generateKeysCmd = &cobra.Command{
	Use:   "generatekeys",
	Short: "Runs lctrld.InitDaemon",
	Long:  ``,
	Run:   generateKeys,
}

var addGenesisAccountsCmd = &cobra.Command{
	Use:   "addgenesisaccounts",
	Short: "Runs lctrld.InitDaemon",
	Long:  ``,
	Run:   addGenesisAccounts,
}

var genTxCmd = &cobra.Command{
	Use:   "gentx",
	Short: "Runs lctrld.GenesisTxs",
	Long:  ``,
	Run:   genTxs,
}

var collectGenTxsCmd = &cobra.Command{
	Use:   "collectgentxs",
	Short: "Runs lctrld.CollectGenesisTxs",
	Long:  ``,
	Run:   collectGenTxs,
}

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(initDaemonCmd)
	payloadCmd.AddCommand(generateKeysCmd)
	payloadCmd.AddCommand(addGenesisAccountsCmd)
	payloadCmd.AddCommand(genTxCmd)
	payloadCmd.AddCommand(collectGenTxsCmd)

}

func initDaemon(cmd *cobra.Command, args []string) {
	lctrld.InitDaemon(settings, args[0])
}

func generateKeys(cmd *cobra.Command, args []string) {
	lctrld.GenerateKeys(settings, args[0])
}

func addGenesisAccounts(cmd *cobra.Command, args []string) {
	lctrld.AddGenesisAccounts(settings, args[0])
}

func genTxs(cmd *cobra.Command, args []string) {
	lctrld.GenesisTxs(settings, args[0])
}

func collectGenTxs(cmd *cobra.Command, args []string) {
	lctrld.CollectGenesisTxs(settings, args[0])
}
