/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/apeunit/evtvzd/pkg/evtvzd"
	"github.com/apeunit/evtvzd/pkg/model"
	"github.com/spf13/cobra"
)

// setupEventCmd represents the setupEvent command
var setupEventCmd = &cobra.Command{
	Use:   "setup-event",
	Short: "Setup a new event",
	Long:  ``,
	Run:   setupEvent,
}

func init() {
	rootCmd.AddCommand(setupEventCmd)
	// token symbol
	setupEventCmd.Flags().StringVarP(&event.TokenSymbol, "token", "t", "", "The token symbol for the Event")
	setupEventCmd.MarkFlagRequired("token")
	// event owner
	setupEventCmd.Flags().StringVarP(&event.Owner, "owner", "o", "", "Set the owner address")
	setupEventCmd.MarkFlagRequired("owner")
	// validators emails
	setupEventCmd.Flags().StringSliceVarP(&event.Validators, "email", "m", []string{}, "Set the validator addresses")
	setupEventCmd.MarkFlagRequired("email")
	// staking variables
	setupEventCmd.Flags().Uint64VarP(&event.Coinbase, "coinbase", "b", model.DefaultCoinbase, "Amount of tokens to be available on the chain and distributed among the validators")
	setupEventCmd.Flags().Uint64VarP(&event.Stake, "stake", "s", model.DefaultStake, "Amount of tokens at stake")
	// provisioning
	setupEventCmd.Flags().StringVar(&event.Provider, "provider", "hetzner", "Provider for provisioning the insfrastructure")
	// TODO add more parameters like: startDate, endDate,
}

var event model.EvtvzE

func setupEvent(cmd *cobra.Command, args []string) {
	fmt.Println("Preparing the environment")
	start := time.Now()

	vc := event.ValidatorsCount()
	tpv := event.Coinbase / uint64(vc)

	fmt.Println("Summary:")
	fmt.Printf("there are %v validators\n", vc)
	fmt.Printf("each with %v stake\n", event.Stake)
	fmt.Printf("and a total coinbase of %s\n", event.FormatAmount(event.Coinbase))
	fmt.Printf("that will be distributed over the %v validators (~ %s each).\n", vc, event.FormatAmount(tpv))
	fmt.Printf("Finally will be deploying %v servers+nodes (1 for each validators) on %s\n", vc, event.Provider)
	fmt.Print("Shall we proceed? [Y/n]:")
	proceed := "Y"
	fmt.Scanln(&proceed)
	if len(proceed) > 0 && strings.ToLower(proceed) != "y" {
		fmt.Println("Ok, nevermind, come back whenever you like")
		return
	}
	fmt.Println("Here we go!!")
	err := evtvzd.DeployEvent(settings, event)
	if err != nil {
		fmt.Println("There was an error, run the command with --debug for more info:", err)
	}
	fmt.Println("Operation completed in", time.Since(start))
}
