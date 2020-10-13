package lctrld

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"

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

// getNodeConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfig/0
func getNodeConfigDir(settings config.Schema, eventID, nodeID string) (configDir string, err error) {
	basePath, err := getConfigDir(settings, eventID)
	if err != nil {
		return
	}
	nodeIDsplit := strings.Split(nodeID, "-")
	return path.Join(basePath, nodeIDsplit[len(nodeIDsplit)-1]), nil
}

// getExtraAccountConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfig/extra_accounts
func getExtraAccountConfigDir(settings config.Schema, eventID, name string) (finalPath string, err error) {
	p, err := getConfigDir(settings, eventID)
	if err != nil {
		return
	}
	return path.Join(p, "extra_accounts", name), nil
}

// DownloadPayloadBinary downloads a copy of the payload binaries to the host
// running lctrld to generate the config files for the provisioned machines
func DownloadPayloadBinary(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (err error) {
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
func InitDaemon(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (*model.EvtvzE, error) {
	log.Infoln("Initializing daemon configs for each node")

	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return nil, err
	}

	_, accounts := evt.Validators()
	for _, acc := range accounts {
		// Make the config directory for the node CLI
		machineConfig := evt.State[acc.Name]
		if acc.Validator {
			nodeConfigDir, err := getNodeConfigDir(settings, evt.ID(), machineConfig.N)
			if err != nil {
				break
			}
			acc.ConfigLocation.DaemonConfigDir = path.Join(nodeConfigDir, "daemon")
			acc.ConfigLocation.CLIConfigDir = path.Join(nodeConfigDir, "cli")
		} else {
			extraAccDir, err := getExtraAccountConfigDir(settings, evt.ID(), acc.Name)
			if err != nil {
				break
			}
			acc.ConfigLocation.CLIConfigDir = extraAccDir
		}

		fmt.Printf("%+v\n", machineConfig)
		fmt.Println("machineConfig.ID", machineConfig.ID())

		args := []string{"init", fmt.Sprintf("%s node %s", acc.Name, machineConfig.ID()), "--home", acc.ConfigLocation.DaemonConfigDir, "--chain-id", evt.ID()}
		out, err := runCommand(settings.EventParams.LaunchPayload.DaemonPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			return nil, err
		}

		args = []string{"tendermint", "show-node-id", "--home", acc.ConfigLocation.DaemonConfigDir}
		out, err = runCommand(settings.EventParams.LaunchPayload.DaemonPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
		}
		machineConfig.TendermintNodeID = strings.TrimSuffix(out, "\n")
	}

	return evt, nil
}

// GenerateKeys generates keys for each genesis account (this includes validator
// accounts). The specific command is gaiacli keys add validatoremail/some other name -o json
// --keyring-backend test --home.... for each node.
func GenerateKeys(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (*model.EvtvzE, error) {
	log.Infoln("Generating keys for validator accounts")

	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return nil, err
	}

	_, validatorAccounts := evt.Validators()
	for _, account := range validatorAccounts {
		args := []string{"keys", "add", account.Name, "-o", "json", "--keyring-backend", "test", "--home", account.ConfigLocation.CLIConfigDir}
		out, err := runCommand(settings.EventParams.LaunchPayload.CLIPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.CLIPath, args, err, out)
			break
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(out), &result)

		account.Address = result["address"].(string)
		account.Mnemonic = result["mnemonic"].(string)

		log.Infof("%s -> %s\n", account.Name, account.Address)
	}

	log.Infoln("Generating keys for non-validator accounts")
	for _, acc := range evt.ExtraAccounts() {
		extraAccDir, err2 := getExtraAccountConfigDir(settings, evt.ID(), acc.Name)
		if err2 != nil {
			return nil, err2
		}

		args := []string{"keys", "add", acc.Name, "-o", "json", "--keyring-backend", "test", "--home", extraAccDir}
		out, err := runCommand(settings.EventParams.LaunchPayload.CLIPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.CLIPath, args, err, out)
			break
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(out), &result)

		acc.Address = result["address"].(string)
		acc.Mnemonic = result["mnemonic"].(string)

		log.Infof("%s -> %s\n", acc.Name, acc.Address)
	}
	return evt, nil
}

// AddGenesisAccounts runs gaiad add-genesis-account with the created addresses
// and default initial balances
func AddGenesisAccounts(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (err error) {
	log.Infoln("Adding accounts to the genesis.json files")

	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return
	}

	for name, state := range evt.State {
		for _, account := range evt.Accounts {
			fmt.Printf("%s %s %s\n", state.ID(), account.Name, account.Address)
			args := []string{"add-genesis-account", account.Address, account.GenesisBalance, "--home", evt.Accounts[name].ConfigLocation.DaemonConfigDir}
			out, err := runCommand(settings.EventParams.LaunchPayload.DaemonPath, args, envVars)
			if err != nil {
				log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
				break
			}
		}
	}

	return
}

// GenesisTxs runs gentx to turn accounts into validator accounts and outputs
// the genesis transactions into a single folder.
func GenesisTxs(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (err error) {
	log.Infoln("Creating genesis transactions to turn accounts into validators")

	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return
	}
	basePath, err := getConfigDir(settings, evt.ID())

	// Ensure that the genesis txs destination directory exists
	outputGenesisTxDir := path.Join(basePath, "genesis_txs")
	if _, err := os.Stat(outputGenesisTxDir); os.IsNotExist(err) {
		os.Mkdir(outputGenesisTxDir, 0755)
	}

	for email, state := range evt.State {
		outputDocument := path.Join(outputGenesisTxDir, fmt.Sprintf("%s.json", state.ID()))
		stakeAmount := strings.Split(evt.Accounts[email].GenesisBalance, ",")

		// Here we assume that last part of genesis_balance is the # of stake tokens
		// launchpayloadd gentx --name v1@email.com --amount 10000stake --home-client ... --keyring-backend test --home ... --output-document ...
		args := []string{"gentx", "--name", email, "--ip", state.Instance.IPAddress, "--amount", stakeAmount[len(stakeAmount)-1], "--home-client", evt.Accounts[email].ConfigLocation.CLIConfigDir, "--keyring-backend", "test", "--home", evt.Accounts[email].ConfigLocation.DaemonConfigDir, "--output-document", outputDocument}
		out, err := runCommand(settings.EventParams.LaunchPayload.DaemonPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			break
		}

	}
	return
}

// CollectGenesisTxs is run on every node's config directory from the single
// directory where the genesis transactions were placed before. In the end, only
// the first node's genesis.josn will be used.
func CollectGenesisTxs(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (err error) {
	log.Infoln("Collecting genesis transactions and writing final genesis.json")

	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return
	}
	basePath, err := getConfigDir(settings, evt.ID())
	if err != nil {
		return
	}
	// Get the first validator/node and use it to generate the genesis.json with all gentxs.
	// firstValidator := evt.Validators[0]

	for name := range evt.State {
		args := []string{"collect-gentxs", "--gentx-dir", path.Join(basePath, "genesis_txs"), "--home", evt.Accounts[name].ConfigLocation.DaemonConfigDir}
		out, err := runCommand(settings.EventParams.LaunchPayload.DaemonPath, args, envVars)
		if err != nil {
			log.Fatalf("%s %s failed with %s, %s\n", settings.EventParams.LaunchPayload.DaemonPath, args, err, out)
			break
		}
	}
	return
}

// EditConfigs edits the config.toml of every node to have the same persistent_peers.
func EditConfigs(settings config.Schema, evt *model.EvtvzE, runCommand CommandRunner) (err error) {
	log.Infoln("Copying node 0's genesis.json to others and setting up p2p.persistent_peers")

	// Although we just generated the genesis.json for every node (makes it
	// easy to debug things) we only need one. Copy node 0's genesis.json to
	// other node folders.
	_, valAccounts := evt.Validators()
	pathToNode0Genesis := path.Join(valAccounts[0].ConfigLocation.DaemonConfigDir, "config/genesis.json")
	for _, valAcc := range valAccounts[1:] {
		node0Genesis, err := os.Open(pathToNode0Genesis)
		otherGenesis := path.Join(valAcc.ConfigLocation.DaemonConfigDir, "config/genesis.json")
		log.Infof("otherGenesis: %s\n", otherGenesis)
		err = os.Remove(otherGenesis)
		if err != nil {
			log.Fatalf("Removing %s failed with %s\n", otherGenesis, err)
			break
		}

		newOtherGenesis, err := os.Create(otherGenesis)
		if err != nil {
			log.Fatalf("Creating a blank %s failed with %s\n", otherGenesis, err)
			break
		}

		written, err := io.Copy(newOtherGenesis, node0Genesis)
		if err != nil {
			log.Fatalf("Copying %s to %s failed with %s\n", pathToNode0Genesis, otherGenesis, err)
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
	for name := range evt.State {
		configPath := path.Join(evt.Accounts[name].ConfigLocation.DaemonConfigDir, "config/config.toml")
		t, err := toml.LoadFile(configPath)
		if err != nil {
			log.Fatalf("Reading toml from file %s failed with %s", configPath, err)
			break
		}
		t.SetPathWithComment([]string{"p2p", "persistent_peers"}, "persistent_peers has been automatically set by lctrld", false, strings.Join(persistentPeerList, ","))

		w, err := os.Create(configPath)
		if err != nil {
			log.Fatalf("Opening file %s in write-mode failed with %s", configPath, err)
			break
		}
		_, err = t.WriteTo(w)
		if err != nil {
			log.Fatalf("Writing TOML to %s failed with %s", configPath, err)
			break
		}
	}
	return
}
