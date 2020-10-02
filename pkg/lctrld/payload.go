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

	"github.com/melbahja/got"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

// getConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfigs
func getConfigDir(settings config.Schema, eventID string) (finalPath string, err error) {
	p, err := evts(settings, eventID)
	if err != nil {
		return
	}
	return path.Join(p, "nodeconfig"), nil
}

// getNodeConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfigs/0
func getNodeConfigDir(settings config.Schema, eventID, nodeID string) (pathDaemon, pathCLI string, err error) {
	basePath, err := getConfigDir(settings, eventID)
	if err != nil {
		return
	}

	nodeIDsplit := strings.Split(nodeID, "-")
	pathDaemon = path.Join(basePath, nodeIDsplit[len(nodeIDsplit)-1], "daemon")
	pathCLI = path.Join(basePath, nodeIDsplit[len(nodeIDsplit)-1], "cli")
	return
}

// DownloadPayloadBinary downloads a copy of the payload binaries to the host
// running lctrld to generate the config files for the provisioned machines
func DownloadPayloadBinary(settings config.Schema, eventID string) (err error) {
	_, cliExistsErr := os.Stat(settings.EventParams.LaunchPayload.CLIPath)
	_, daemonExistsErr := os.Stat(settings.EventParams.LaunchPayload.DaemonPath)
	if os.IsNotExist(cliExistsErr) || os.IsNotExist(daemonExistsErr) {
		binFile := bin(settings, "payloadBinaries.zip")
		log.Infof("downloading payload binaries from %s to %s", settings.EventParams.LaunchPayload.BinaryURL, binFile)
		g := got.New()
		err = g.Download(settings.EventParams.LaunchPayload.BinaryURL, binFile)
		if err != nil {
			return
		}

		_, err = runCommand("unzip", []string{"-d", bin(settings, ""), "-o", binFile}, []string{})
		if err != nil {
			return
		}

		err = os.Remove(binFile)
		if err != nil {
			return
		}
	}
	return nil
}

// InitDaemon runs gaiad init burnerchain --home
// state.DaemonConfigDir
// and gaiad tendermint show-node-id
func InitDaemon(settings config.Schema, eventID string) (err error) {
	log.Infoln("Initializing daemon configs for each node")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	for email, state := range evt.State {
		// Make the config directory for the node CLI
		pathDaemon, pathCLI, err := getNodeConfigDir(settings, eventID, state.ID)
		if err != nil {
			break
		}
		state.DaemonConfigDir = pathDaemon
		state.CLIConfigDir = pathCLI

		args := []string{"init", fmt.Sprintf("%s node %s", email, state.ID), "--home", state.DaemonConfigDir, "--chain-id", evt.ID()}
		cmd := exec.Command(settings.EventParams.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			return err
		}

		args = []string{"tendermint", "show-node-id", "--home", state.DaemonConfigDir}
		cmd = exec.Command(settings.EventParams.LaunchPayload.DaemonPath, args...)
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
		}
		state.TendermintNodeID = strings.TrimSuffix(string(out), "\n")
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}

// GenerateKeys generates keys for each genesis account (this includes validator
// accounts). The specific command is gaiacli keys add validatoremail/some other name -o json
// --keyring-backend test --home.... for each node.
func GenerateKeys(settings config.Schema, eventID string) (err error) {
	log.Infoln("Generating keys")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	_, validatorAccounts := evt.Validators()
	for _, account := range validatorAccounts {
		args := []string{"keys", "add", account.Name, "-o", "json", "--keyring-backend", "test", "--home", evt.State[account.Name].CLIConfigDir}
		cmd := exec.Command(settings.EventParams.LaunchPayload.CLIPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.CLIPath, args, err, out)
			break
		}

		var result map[string]interface{}
		json.Unmarshal(out, &result)

		account.Address = result["address"].(string)
		account.Mnemonic = result["mnemonic"].(string)

		log.Infof("%s -> %s\n", account.Name, account.Address)
	}

	err = storeEvent(settings, evt)
	if err != nil {
		return
	}
	return
}

// AddGenesisAccounts runs gaiad add-genesis-account with the created addresses
// and default initial balances
func AddGenesisAccounts(settings config.Schema, eventID string) (err error) {
	log.Infoln("Adding accounts to the genesis.json files")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	_, validatorAccounts := evt.Validators()
	for _, state := range evt.State {
		for _, account := range validatorAccounts {
			args := []string{"add-genesis-account", account.Address, account.GenesisBalance, "--home", state.DaemonConfigDir}
			cmd := exec.Command(settings.EventParams.LaunchPayload.DaemonPath, args...)
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
				break
			}
		}
	}

	return
}

// GenesisTxs runs gentx to turn accounts into validator accounts and outputs
// the genesis transactions into a single folder.
func GenesisTxs(settings config.Schema, eventID string) (err error) {
	log.Infoln("Creating genesis transactions to turn accounts into validators")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}
	basePath, err := getConfigDir(settings, eventID)

	// Ensure that the genesis txs destination directory exists
	outputGenesisTxDir := path.Join(basePath, "genesis_txs")
	if _, err := os.Stat(outputGenesisTxDir); os.IsNotExist(err) {
		os.Mkdir(outputGenesisTxDir, 0755)
	}

	for email, state := range evt.State {
		outputDocument := path.Join(outputGenesisTxDir, fmt.Sprintf("%s.json", state.ID))
		stakeAmount := strings.Split(evt.Accounts[email].GenesisBalance, ",")

		// Here we assume that last part of genesis_balance is the # of stake tokens
		// launchpayloadd gentx --name v1@email.com --amount 10000stake --home-client ... --keyring-backend test --home ... --output-document ...
		args := []string{"gentx", "--name", email, "--ip", state.Instance.IPAddress, "--amount", stakeAmount[len(stakeAmount)-1], "--home-client", state.CLIConfigDir, "--keyring-backend", "test", "--home", state.DaemonConfigDir, "--output-document", outputDocument}
		cmd := exec.Command(settings.EventParams.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			break
		}

	}
	return
}

// CollectGenesisTxs is run on every node's config directory from the single
// directory where the genesis transactions were placed before. In the end, only
// the first node's genesis.josn will be used.
func CollectGenesisTxs(settings config.Schema, eventID string) (err error) {
	log.Infoln("Collecting genesis transactions and writing final genesis.json")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}
	basePath, err := getConfigDir(settings, eventID)
	if err != nil {
		return
	}
	// Get the first validator/node and use it to generate the genesis.json with all gentxs.
	// firstValidator := evt.Validators[0]

	for _, state := range evt.State {
		args := []string{"collect-gentxs", "--gentx-dir", path.Join(basePath, "genesis_txs"), "--home", state.DaemonConfigDir}
		cmd := exec.Command(settings.EventParams.LaunchPayload.DaemonPath, args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			break
		}
	}
	return
}

// EditConfigs edits the config.toml of every node to have the same persistent_peers.
func EditConfigs(settings config.Schema, eventID string) (err error) {
	log.Infoln("Copying node 0's genesis.json to others and setting up p2p.persistent_peers")
	evt, err := loadEvent(settings, eventID)
	if err != nil {
		return
	}

	// Although we just generated the genesis.json for every node (makes it
	// easy to debug things) we only need one. Copy node 0's genesis.json to
	// other node folders.
	validatorNames, _ := evt.Validators()
	pathToNode0Genesis := path.Join(evt.State[validatorNames[0]].DaemonConfigDir, "config/genesis.json")
	for _, validator := range validatorNames[1:] {
		node0Genesis, err := os.Open(pathToNode0Genesis)
		otherGenesis := path.Join(evt.State[validator].DaemonConfigDir, "config/genesis.json")
		log.Infof("otherGenesis: %s\n", otherGenesis)
		err = os.Remove(otherGenesis)
		if err != nil {
			log.Errorf("Removing %s failed with %s\n", otherGenesis, err)
			break
		}

		newOtherGenesis, err := os.Create(otherGenesis)
		if err != nil {
			log.Errorf("Creating a blank %s failed with %s\n", newOtherGenesis, err)
			break
		}

		written, err := io.Copy(newOtherGenesis, node0Genesis)
		if err != nil {
			log.Errorf("Copying %s to %s failed with %s\n", node0Genesis, newOtherGenesis, err)
			break
		}
		log.Debugf("Copied %v bytes to %s", written, otherGenesis)
	}

	// Build the persistent peer list.
	persistentPeerList := []string{}
	for email, state := range evt.State {
		fmt.Printf("%s's node is %s \n", email, state.TendermintPeerNodeID())
		persistentPeerList = append(persistentPeerList, state.TendermintPeerNodeID())
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
