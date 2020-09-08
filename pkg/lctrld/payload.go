package lctrld

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/pelletier/go-toml"
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

	dmBin := dmBin(settings)
	// set the path to find the executable
	evnVars, err := dockerEnv(settings, evt)

	for email, state := range evt.State {
		ip, err := runCommand(dmBin, []string{"ip", state.ID}, evnVars)
		if err != nil {
			log.Error(err)
			break
		}
		state.Instance.IPAddress = string(ip)

		fmt.Println("Node owner is", email)
		fmt.Println("Node IP is", state.Instance.IPAddress)

		// Make the config directory for the node CLI
		pathDaemon, pathCLI, err := getConfigDir(settings, eventID, state.ID)
		if err != nil {
			break
		}
		state.DaemonConfigDir = pathDaemon
		state.CLIConfigDir = pathCLI

		args := []string{"init", fmt.Sprintf("%s node %s", email, state.ID), "--home", state.DaemonConfigDir, "--chain-id", evt.ID()}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			return err
		}

		args = []string{"tendermint", "show-node-id", "--home", state.DaemonConfigDir}
		cmd = exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
		}
		state.TendermintNodeID = strings.TrimSuffix(string(out), "\n")

		fmt.Printf("Final State: %+v\n", state)
	}

	fmt.Printf("evt now looks like: %+v\n", evt)
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
	addresses := []string{}
	for _, state := range evt.State {
		addresses = append(addresses, state.Account.Address)
	}
	fmt.Println("addresses", addresses)

	for _, state := range evt.State {
		for _, addr := range addresses {
			args := []string{"add-genesis-account", addr, evt.GenesisDeclaration(), "--home", state.DaemonConfigDir}
			cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
				break
			}
		}
	}

	return
}

func GenesisTxs(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}
	evtDir, err := evts(settings, evt.ID())
	if err != nil {
		return
	}

	// Ensure that the genesis txs destination directory exists
	outputGenesisTxDir := path.Join(evtDir, "genesis_txs")
	if _, err := os.Stat(outputGenesisTxDir); os.IsNotExist(err) {
		os.Mkdir(outputGenesisTxDir, 0755)
	}

	for email, state := range evt.State {
		outputDocument := path.Join(outputGenesisTxDir, fmt.Sprintf("%s.json", state.ID))
		stakeAmount := strings.Split(state.Account.GenesisBalance, ",")

		args := []string{"gentx", "--name", email, "--amount", stakeAmount[len(stakeAmount)-1], "--home-client", state.CLIConfigDir, "--keyring-backend", "test", "--home", state.DaemonConfigDir, "--output-document", outputDocument}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			break
		}

	}
	return
}

func CollectGenesisTxs(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}
	evtDir, err := evts(settings, evt.ID())
	if err != nil {
		return
	}
	// Get the first validator/node and use it to generate the genesis.json with all gentxs.
	// firstValidator := evt.Validators[0]

	for _, state := range evt.State {
		args := []string{"collect-gentxs", "--gentx-dir", path.Join(evtDir, "genesis_txs"), "--home", state.DaemonConfigDir}
		cmd := exec.Command(settings.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.LaunchPayload.DaemonPath, args, err, out)
			break
		}
	}

	// Although we just generated the genesis.json for every node (makes it
	// easy to debug things) we only need one. Copy node 0's genesis.json to
	// other node folders.
	otherValidators := evt.Validators[1:]
	pathToNode0Genesis := path.Join(evt.State[evt.Validators[0]].DaemonConfigDir, "config/genesis.json")
	node0Genesis, err := os.Open(pathToNode0Genesis)
	for _, validator := range otherValidators {
		otherGenesis := path.Join(evt.State[validator].DaemonConfigDir, "config/genesis.json")
		err := os.Remove(otherGenesis)
		if err != nil {
			log.Errorf("Removing %s failed with %s\n", otherGenesis, err)
			break
		}

		newOtherGenesis, err := os.Create(otherGenesis)
		if err != nil {
			log.Errorf("Creating a blank %s failed with %s\n", newOtherGenesis, err)
			break
		}

		_, err = io.Copy(newOtherGenesis, node0Genesis)
		if err != nil {
			log.Errorf("Copying %s to %s failed with %s\n", node0Genesis, newOtherGenesis, err)
			break
		}
	}
	return
}

func EditConfigs(settings config.Schema, eventID string) (err error) {
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	// Build the persistent peer list
	persistentPeerList := []string{}
	for email, state := range evt.State {
		fmt.Printf("%s has tendermint-node-id %s and IP %s\n", email, state.TendermintNodeID, state.Instance.IPAddress)
		persistentPeerList = append(persistentPeerList, fmt.Sprintf("%s@%s:26656", state.TendermintNodeID, state.Instance.IPAddress))
	}

	// Insert the persistent peer list into each node's config.toml
	for _, state := range evt.State {
		configPath := path.Join(state.DaemonConfigDir, "config/config.toml")
		t, err := toml.LoadFile(configPath)
		if err != nil {
			log.Errorf("Reading toml from file %s failed with %s", configPath, err)
			break
		}
		t.SetPathWithComment([]string{"p2p", "persistent_peers"}, "persistent_peers has been automatically set by lctrld", false, strings.Join(persistentPeerList, ","))

		w, err := os.Create(configPath)
		if err != nil {
			log.Errorf("Opening file %s in write-mode failed with %s", configPath, err)
			break
		}
		_, err = t.WriteTo(w)
		if err != nil {
			log.Errorf("Writing TOML to %s failed with %s", configPath, err)
			break
		}
	}
	return
}
