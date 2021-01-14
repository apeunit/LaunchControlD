package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/apeunit/LaunchControlD/pkg/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// eventsCmd represents the events command
var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Manage events",
	Long:  ``,
}
var provider string

func init() {
	rootCmd.AddCommand(eventsCmd)

	eventsCmd.AddCommand(setupEventCmd)
	setupEventCmd.Flags().StringVar(&provider, "provider", "hetzner", "Provider for provisioning the insfrastructure")

	eventsCmd.AddCommand(tearDownEventCmd)

	eventsCmd.AddCommand(listEventCmd)
	listEventCmd.Flags().BoolVar(&verbose, "verbose", false, "Print more details")

	eventsCmd.AddCommand(retryEventCmd)
}

var verbose bool

// setupEventCmd represents the setupEvent command
var setupEventCmd = &cobra.Command{
	Use:   "new <eventrequest YAML file>",
	Short: "Setup a new event",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  setupEvent,
}

func setupEvent(cmd *cobra.Command, args []string) (err error) {
	start := time.Now()

	evtRequest, err := model.LoadEventRequestFromFile(args[0])
	if err != nil {
		return
	}
	evtRequest.PayloadLocation = settings.DefaultPayloadLocation

	evt := model.NewEvent(evtRequest.TokenSymbol, evtRequest.Owner, provider, evtRequest.GenesisAccounts, evtRequest.PayloadLocation)
	vc := evt.ValidatorsCount()

	log.Debugf("%#v\n", settings)
	log.Debugf("%#v\n", evtRequest)
	log.Debugf("%#v\n", evt)
	fmt.Println("Summary:")
	_, validatorAccounts := evt.Validators()
	for _, acc := range validatorAccounts {
		fmt.Printf("Validator %s has initial balance of %+v\n", acc.Name, acc.GenesisBalance)
	}
	fmt.Printf("Including other accounts, the genesis account state is:\n")
	for k, v := range evt.Accounts {
		fmt.Printf("%s: %+v\n", k, v)
	}
	fmt.Printf("Finally will be deploying %v servers+nodes (1 for each validators) on %s\n", vc, evt.Provider)
	fmt.Print("Shall we proceed? [Y/n]:")
	proceed := "Y"
	fmt.Scanln(&proceed)
	if len(proceed) > 0 && strings.ToLower(proceed) != "y" {
		fmt.Println("Ok, nevermind, come back whenever you like")
		return
	}
	fmt.Println("Here we go!!")
	err = lctrld.CreateEvent(settings, evt)
	if err != nil {
		log.Error("There was an error, run the command with --debug for more info:", err)
		return err
	}

	dmc := lctrld.NewDockerMachineConfig(settings, evt.ID())
	err = lctrld.Provision(settings, evt, lctrld.RunCommand, dmc)
	if err != nil {
		log.Error("There was an error, run the command with --debug for more info:", err)
		return err
	}
	err = lctrld.StoreEvent(settings, evt)
	if err != nil {
		log.Error("There was a problem saving the updated Event", err)
		return err
	}
	fmt.Println("Operation completed in", time.Since(start))
	return nil
}

// tearDownEventCmd represents the tearDownEvent command
var tearDownEventCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Destroy the resources associated to an event",
	Long:  ``,
	Args:  cobra.ExactArgs(1),
	RunE:  tearDownEvent,
}

func tearDownEvent(cmd *cobra.Command, args []string) (err error) {
	fmt.Println("Teardown Event")
	fmt.Println("Event ID is", args[0])
	start := time.Now()
	evt, err := lctrld.LoadEvent(settings, args[0])
	if err != nil {
		log.Error("There was an error shutting down the event: ", err)
		return err
	}
	err = lctrld.DestroyEvent(settings, evt, lctrld.RunCommand)
	if err != nil {
		log.Error("There was an error shutting down the event: ", err)
		return err
	}
	fmt.Println("Operation completed in", time.Since(start))
	return nil
}

// listEventCmd represents the tearDownEvent command
var listEventCmd = &cobra.Command{
	Use:   "list",
	Short: "list available events",
	Long:  ``,
	Run:   listEvent,
}

func listEvent(cmd *cobra.Command, args []string) {
	fmt.Println("List events")
	start := time.Now()
	events, err := lctrld.ListEvents(settings)
	if err != nil {
		fmt.Println("There was an error shutting down the event: ", err)
	}
	for _, evt := range events {
		fmt.Println("Event", evt.ID(), "owner:", evt.Owner, "with", evt.ValidatorsCount(), "validators")
		if verbose {
			lctrld.InspectEvent(settings, &evt, lctrld.RunCommand)
		}
	}
	fmt.Println("Operation completed in", time.Since(start))
}

// listEventCmd represents the tearDownEvent command
var retryEventCmd = &cobra.Command{
	Use:   "retry",
	Short: "rereads docker-machine's .docker/machine/machines/<name>/config.json into event.json, in case a human fixed whatever went wrong",
	Long:  "rereads docker-machine's .docker/machine/machines/<name>/config.json into event.json, in case a human fixed whatever went wrong",
	Args:  cobra.ExactArgs(1),
	RunE:  retryEvent,
}

func retryEvent(cmd *cobra.Command, args []string) (err error) {
	evt, err := lctrld.LoadEvent(settings, args[0])
	if err != nil {
		return
	}
	dmc := lctrld.NewDockerMachineConfig(settings, evt.ID())
	evt2, err := lctrld.RereadDockerMachineInfo(settings, evt, dmc)
	if err != nil {
		return
	}
	err = lctrld.StoreEvent(settings, evt2)
	if err != nil {
		return
	}
	validatorNames, _ := evt2.Validators()
	for _, v := range validatorNames {
		log.Infof("Updated info for %s: %#v\n", v, evt2.State[v])
	}
	if err != nil {
		log.Error("There was a problem saving the updated Event", err)
		return
	}
	return
}
