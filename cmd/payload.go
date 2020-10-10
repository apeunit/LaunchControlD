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
	Use:   "setup EVENTID",
	Short: "Does everything to initialize a Cosmos-SDK based payload for EVENTID",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  setupChain,
}

var deployCmd = &cobra.Command{
	Use:   "deploy EVENTID",
	Short: "Tells the provisioned machines to run the dockerized payload for EVENTID",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  deploy,
}

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(setupChainCmd)
	payloadCmd.AddCommand(deployCmd)
}

func setupChain(cmd *cobra.Command, args []string) (err error) {
	err = lctrld.DownloadPayloadBinary(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.InitDaemon(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.GenerateKeys(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.AddGenesisAccounts(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.GenesisTxs(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.CollectGenesisTxs(settings, args[0])
	if err != nil {
		return
	}
	err = lctrld.EditConfigs(settings, args[0])
	if err != nil {
		return
	}
	return
}

func deploy(cmd *cobra.Command, args []string) (err error) {
	evt, err := lctrld.LoadEvent(settings, args[0])
	if err != nil {
		return err
	}
	err = lctrld.DeployPayload(settings, evt)
	return
}
