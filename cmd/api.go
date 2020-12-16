package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/apeunit/LaunchControlD/pkg/lctrld"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "HTTP REST API to remotely control lctrld",
	Long:  ``,
	RunE:  runAPI,
}

func init() {
	rootCmd.AddCommand(apiCmd)
}

func runAPI(cmd *cobra.Command, args []string) (err error) {
	router := mux.NewRouter()

	router.HandleFunc("/events", listEvents).Methods("GET")
	router.HandleFunc("/event", createEvent).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
	return nil
}

func listEvents(w http.ResponseWriter, r *http.Request) {
	events, _ := lctrld.ListEvents(settings)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	var e model.EventRequest
	json.NewDecoder(r.Body).Decode(&e)
	log.Printf("%#v\n", e)
	event := model.NewEvent(e.TokenSymbol, e.Owner, "virtualbox", e.GenesisAccounts, e.PayloadLocation)
	err := lctrld.CreateEvent(settings, event)
	log.Printf("Creating event %#v\n", event)
	if err != nil {
		errMsg := fmt.Sprintf("{\"error\": \"%s\"}", err)
		w.Write([]byte(errMsg))
	}

	dmc := lctrld.NewDockerMachineConfig(settings, event.ID())
	err = lctrld.Provision(settings, event, lctrld.RunCommand, dmc)
	if err != nil {
		errMsg := fmt.Sprintf("{\"error\": \"%s\"}", err)
		w.Write([]byte(errMsg))
	}

	err = lctrld.StoreEvent(settings, event)
	if err != nil {
		errMsg := fmt.Sprintf("{\"error\": \"%s\"}", err)
		w.Write([]byte(errMsg))
	}

	successMsg := fmt.Sprintf("{\"id\": \"%s\"}", event.ID())
	w.Write([]byte(successMsg))
}
