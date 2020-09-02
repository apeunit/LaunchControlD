package lctrld

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	log "github.com/sirupsen/logrus"
)

func getConfigDir(settings config.Schema, eventID, nodeID string) (pathDaemon, pathCLI string, err error) {
	p, err := evts(settings, eventID)
	if err != nil {
		return
	}
	nodeIDsplit := strings.Split(nodeID, "-")
	pathDaemon = path.Join(p, nodeIDsplit[len(nodeIDsplit)-1], "daemon")
	pathCLI = path.Join(p, nodeIDsplit[len(nodeIDsplit)-1], "cli")
	return
}

// InitDaemon runs gaiad init burnerchain --home
// state.DaemonConfigDir
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
		state.DaemonConfigDir = pathDaemon
		state.CLIConfigDir = pathCLI
		fmt.Printf("%+v\n", evt.State)

		args := []string{"init", "burnerchain", "--home", state.DaemonConfigDir}
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
		state.TendermintNodeID = strings.TrimSuffix(string(out), "\n")

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

		args := []string{"keys", "add", email, "-o", "json", "--keyring-backend", "test", "--home", state.CLIConfigDir}
		cmd := exec.Command(settings.LaunchPayload.CLIPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.CLIPath, args, err, out)
			break
		}

		var result map[string]interface{}
		json.Unmarshal(out, &result)
		state.Account.Name = email
		state.Account.Address = result["address"].(string)
		state.Account.Mnemonic = result["mnemonic"].(string)
		state.Account.GenesisBalance = evt.GenesisDeclaration()
		fmt.Printf("%s -> %s with mnemonic \"%s\"\n", email, state.Account.Address, state.Account.Mnemonic)
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}

func AddGenesisAccounts(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	for _, state := range evt.State {
		args := []string{"add-genesis-account", state.Account.Address, evt.GenesisDeclaration(), "--home", state.DaemonConfigDir}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			break
		}
	}

	return
}

func GenesisTxs(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	for email, state := range evt.State {
		stakeAmount := strings.Split(state.Account.GenesisBalance, ",")

		args := []string{"gentx", "--name", email, "--amount", stakeAmount[len(stakeAmount)-1], "--home-client", state.CLIConfigDir, "--keyring-backend", "test", "--home", state.DaemonConfigDir}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			break
		}
	}
	return
}
