package lctrld

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path"

	"github.com/apeunit/LaunchControlD/pkg/config"
	log "github.com/sirupsen/logrus"
)

func getConfigDir(settings config.Schema, eventID, nodeID string) (pathDaemon, pathCLI string, err error) {
	p, err := evts(settings, eventID)
	if err != nil {
		return
	}

	pathDaemon = path.Join(p, nodeID, "daemon")
	pathCLI = path.Join(p, nodeID, "cli")
	return
}

// InitDaemon runs gaiad init burnerchain --home
// state.PayloadConfig.DaemonConfigDir
// and gaiad tendermint show-node-id
func InitDaemon(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	for email, state := range evt.State {
		fmt.Println("Node owner is", email)
		fmt.Println("Node IP is", state.Instance.IPAddress)

		// Make the config directory for the node CLI
		pathDaemon, pathCLI, err := getConfigDir(settings, eventID, state.ID)
		if err != nil {
			break
		}
		state.PayloadConfig.DaemonConfigDir = pathDaemon
		state.PayloadConfig.CLIConfigDir = pathCLI
		fmt.Printf("%+v\n", evt.State)

		args := []string{"init", "burnerchain", "--home", state.PayloadConfig.DaemonConfigDir}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			return err
		}

		args = []string{"tendermint", "show-node-id"}
		cmd = exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
		}
		state.PayloadConfig.TendermintNodeID = string(out)

		fmt.Printf("State 2: %+v\n", state)
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}

// GenerateKeys generates keys for each validator. The specific command is
// gaiacli keys add validatoremail -o json --keyring-backend test --home.... for
// each node.
func GenerateKeys(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	for email, state := range evt.State {
		fmt.Println("Node owner is", email)
		fmt.Println("Node IP is", state.Instance.IPAddress)

		args := []string{"keys", "add", email, "-o", "json", "--keyring-backend", "test", "--home", state.PayloadConfig.CLIConfigDir}
		cmd := exec.Command(settings.LaunchPayload.CLIPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.CLIPath, args, err, out)
			break
		}

		var result map[string]interface{}
		json.Unmarshal(out, &result)
		state.PayloadConfig.Account.Address = result["address"].(string)
		state.PayloadConfig.Account.Mnemonic = result["mnemonic"].(string)
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}
