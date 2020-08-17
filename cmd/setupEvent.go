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
	setupEventCmd.Flags().StringVarP(&tokenSymbol, "token", "t", "EVZ", "The token symbol for the Event")
	setupEventCmd.MarkFlagRequired("token")
	// validators emails
	setupEventCmd.Flags().StringSliceVarP(&validatorEmails, "email", "m", []string{}, "Set the validator addresses")
	setupEventCmd.MarkFlagRequired("email")
	// staking variables
	setupEventCmd.Flags().Uint64VarP(&coinbase, "coinbase", "b", 1000000000, "Amount of tokens to be available on the chain and distributed among the validators")
	setupEventCmd.Flags().Uint64VarP(&stake, "stake", "s", 1000000, "Amount of tokens at stake")
	// provisioning
	setupEventCmd.Flags().StringSliceVar(&providers, "providers", []string{"hetzner"}, "Providers for provisioning the insfrastructure")
	// TODO add more parameters like: startDate, endDate,
}

var (
	tokenSymbol     string
	coinbase        uint64
	stake           uint64
	validatorEmails []string
	providers       []string
)

func setupEvent(cmd *cobra.Command, args []string) {
	fmt.Println("Preparing the environment")

	vc := len(validatorEmails)
	tpv := coinbase / uint64(vc)

	fmt.Println("Summary:")
	fmt.Printf("there are %v validators\n", vc)
	fmt.Printf("each with %v stake\n", stake)
	fmt.Printf("and a total coinbase of %v%s\n", coinbase, tokenSymbol)
	fmt.Printf("that will be distributed over the %v validators (~ %v%s each).\n", vc, tpv, tokenSymbol)
	fmt.Printf("Finally will be deploying %v servers+nodes (1 for each validators) on %s\n", vc, strings.Join(providers, ", "))
	fmt.Print("Shall we proceed? [Y/n]:")
	proceed := "Y"
	fmt.Scanln(&proceed)
	if len(proceed) > 0 && strings.ToLower(proceed) != "y" {
		fmt.Println("Ok, nevermind, come back whenever you like")
		return
	}
	fmt.Println("Here we go!!")
}
