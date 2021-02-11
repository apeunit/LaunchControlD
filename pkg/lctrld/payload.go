package lctrld

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"

	"github.com/melbahja/got"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
)

// DownloadPayloadBinary downloads a copy of the payload binaries to the host
// running lctrld to generate the config files for the provisioned machines
func DownloadPayloadBinary(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	_, cliExistsErr := os.Stat(evt.Payload.CLIPath)
	_, daemonExistsErr := os.Stat(evt.Payload.DaemonPath)
	if os.IsNotExist(cliExistsErr) || os.IsNotExist(daemonExistsErr) {
		binFile := settings.Bin("payloadBinaries.zip")
		log.Infof("downloading payload binaries from %s to %s", evt.Payload.BinaryURL, binFile)
		g := got.New()
		err = g.Download(evt.Payload.BinaryURL, binFile)
		if err != nil {
			return
		}

		_, err = runCommand([]string{"unzip", "-d", settings.Bin(""), "-o", binFile}, []string{})
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
func InitDaemon(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (*model.Event, error) {
	log.Infoln("Initializing daemon configs for each node")

	envVars := utils.BuildEnvVars(settings)

	_, accounts := evt.Validators()
	for _, acc := range accounts {
		// Make the config directory for the node CLI
		machineConfig := evt.State[acc.Name]
		if acc.Validator {
			nodeConfigDir, err := settings.NodeConfigDir(evt.ID(), machineConfig.N)
			if err != nil {
				break
			}
			acc.ConfigLocation.DaemonConfigDir = path.Join(nodeConfigDir, "daemon")
			acc.ConfigLocation.CLIConfigDir = path.Join(nodeConfigDir, "cli")
		} else {
			extraAccDir, err := settings.ExtraAccountConfigDir(evt.ID(), acc.Name)
			if err != nil {
				break
			}
			acc.ConfigLocation.CLIConfigDir = extraAccDir
		}

		command := []string{evt.Payload.DaemonPath, "init", fmt.Sprintf("%s node %s", acc.Name, machineConfig.ID()), "--home", acc.ConfigLocation.DaemonConfigDir, "--chain-id", evt.ID()}
		out, err := runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.DaemonPath, command, err, out)
			return nil, err
		}

		command = []string{evt.Payload.DaemonPath, "tendermint", "show-node-id", "--home", acc.ConfigLocation.DaemonConfigDir}
		out, err = runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.DaemonPath, command, err, out)
			return nil, err
		}
		machineConfig.TendermintNodeID = strings.TrimSuffix(out, "\n")
	}

	return evt, nil
}

// GenerateKeys generates keys for each genesis account (this includes validator
// accounts). The specific command is gaiacli keys add validatoremail/some other name -o json
// --keyring-backend test --home.... for each node.
func GenerateKeys(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (*model.Event, error) {
	log.Infoln("Generating keys for validator accounts")

	envVars := utils.BuildEnvVars(settings)

	_, validatorAccounts := evt.Validators()
	for _, account := range validatorAccounts {
		command := []string{evt.Payload.CLIPath, "keys", "add", account.Name, "-o", "json", "--keyring-backend", "test", "--home", account.ConfigLocation.CLIConfigDir}
		out, err := runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.CLIPath, command, err, out)
			return nil, err
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(out), &result)

		account.Address = result["address"].(string)
		account.Mnemonic = result["mnemonic"].(string)

		log.Infof("%s -> %s\n", account.Name, account.Address)
	}

	log.Infoln("Generating keys for non-validator accounts")
	for _, acc := range evt.ExtraAccounts() {
		extraAccDir, err2 := settings.ExtraAccountConfigDir(evt.ID(), acc.Name)
		if err2 != nil {
			return nil, err2
		}

		command := []string{evt.Payload.CLIPath, "keys", "add", acc.Name, "-o", "json", "--keyring-backend", "test", "--home", extraAccDir}
		out, err := runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.CLIPath, command, err, out)
			return nil, err
		}

		var result map[string]interface{}
		json.Unmarshal([]byte(out), &result)

		acc.Address = result["address"].(string)
		acc.Mnemonic = result["mnemonic"].(string)
		acc.ConfigLocation.CLIConfigDir = extraAccDir

		log.Infof("%s -> %s\n", acc.Name, acc.Address)
	}
	return evt, nil
}

// AddGenesisAccounts runs gaiad add-genesis-account with the created addresses
// and default initial balances
func AddGenesisAccounts(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	log.Infoln("Adding accounts to the genesis.json files")

	envVars := utils.BuildEnvVars(settings)

	for name := range evt.State {
		for _, account := range evt.Accounts {
			command := []string{evt.Payload.DaemonPath, "add-genesis-account", account.Address, account.GenesisBalance, "--home", evt.Accounts[name].ConfigLocation.DaemonConfigDir}
			out, err := runCommand(command, envVars)
			if err != nil {
				log.Errorf("%s %s failed with %s, %s\n", evt.Payload.DaemonPath, command, err, out)
				return err
			}
		}
	}

	return
}

// GenesisTxs runs gentx to turn accounts into validator accounts and outputs
// the genesis transactions into a single folder.
func GenesisTxs(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	log.Infoln("Creating genesis transactions to turn accounts into validators")

	envVars := utils.BuildEnvVars(settings)
	basePath, err := settings.ConfigDir(evt.ID())
	if err != nil {
		return
	}

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
		command := []string{evt.Payload.DaemonPath, "gentx", "--name", email, "--ip", state.Instance.IPAddress, "--amount", stakeAmount[len(stakeAmount)-1], "--home-client", evt.Accounts[email].ConfigLocation.CLIConfigDir, "--keyring-backend", "test", "--home", evt.Accounts[email].ConfigLocation.DaemonConfigDir, "--output-document", outputDocument}
		out, err := runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.DaemonPath, command, err, out)
			return err
		}

	}
	return
}

// CollectGenesisTxs is run on every node's config directory from the single
// directory where the genesis transactions were placed before. In the end, only
// the first node's genesis.json will be used.
func CollectGenesisTxs(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	log.Infoln("Collecting genesis transactions and writing final genesis.json")

	envVars := utils.BuildEnvVars(settings)
	basePath, err := settings.ConfigDir(evt.ID())
	if err != nil {
		return
	}
	// Get the first validator/node and use it to generate the genesis.json with all gentxs.
	// firstValidator := evt.Validators[0]

	for name := range evt.State {
		command := []string{evt.Payload.DaemonPath, "collect-gentxs", "--gentx-dir", path.Join(basePath, "genesis_txs"), "--home", evt.Accounts[name].ConfigLocation.DaemonConfigDir}
		out, err := runCommand(command, envVars)
		if err != nil {
			log.Errorf("%s %s failed with %s, %s\n", evt.Payload.DaemonPath, command, err, out)
			return err
		}
	}
	return
}

// EditConfigs edits the config.toml of every node to have the same persistent_peers.
func EditConfigs(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	log.Infoln("Copying node 0's genesis.json to others and setting up p2p.persistent_peers")

	// Although we just generated the genesis.json for every node (makes it
	// easy to debug things) we only need one. Copy node 0's genesis.json to
	// other node folders.
	_, valAccounts := evt.Validators()
	pathToNode0Genesis := path.Join(valAccounts[0].ConfigLocation.DaemonConfigDir, "config/genesis.json")
	for _, valAcc := range valAccounts[1:] {
		node0Genesis, err := os.Open(pathToNode0Genesis)
		if err != nil {
			log.Errorf("cannot open file genesis descriptor: %s: %v", pathToNode0Genesis, err)
			return err
		}
		otherGenesis := path.Join(valAcc.ConfigLocation.DaemonConfigDir, "config/genesis.json")
		log.Infof("otherGenesis: %s\n", otherGenesis)
		err = os.Remove(otherGenesis)
		if err != nil {
			log.Errorf("Removing %s failed with %s\n", otherGenesis, err)
			return err
		}

		newOtherGenesis, err := os.Create(otherGenesis)
		if err != nil {
			log.Errorf("Creating a blank %s failed with %s\n", otherGenesis, err)
			return err
		}

		written, err := io.Copy(newOtherGenesis, node0Genesis)
		if err != nil {
			log.Errorf("Copying %s to %s failed with %s\n", pathToNode0Genesis, otherGenesis, err)
			return err
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
	// Don't create blocks if there are no txs (to save space when chain is idle)
	for name := range evt.State {
		configPath := path.Join(evt.Accounts[name].ConfigLocation.DaemonConfigDir, "config/config.toml")
		t, err := toml.LoadFile(configPath)
		if err != nil {
			log.Errorf("Reading toml from file %s failed with %s", configPath, err)
			return err
		}
		t.SetPathWithComment([]string{"p2p", "persistent_peers"}, "persistent_peers has been automatically set by lctrld", false, strings.Join(persistentPeerList, ","))
		t.SetPathWithComment([]string{"rpc", "laddr"}, "laddr has been automatically set by lctrld", false, "tcp://0.0.0.0:26657")
		t.SetPathWithComment([]string{"consensus", "create_empty_blocks"}, "Don't create blocks if there are no txs: automatically set by lctrld", false, false)

		w, err := os.Create(configPath)
		if err != nil {
			log.Errorf("Opening file %s in write-mode failed with %s", configPath, err)
			return err
		}
		_, err = t.WriteTo(w)
		if err != nil {
			log.Errorf("Writing TOML to %s failed with %s", configPath, err)
			return err
		}
	}
	return
}

// GenerateFaucetConfig generates a configuration for the faucet given what it knows about the event
func GenerateFaucetConfig(settings *config.Schema, evt *model.Event, runCommand cmdrunner.CommandRunner) (err error) {
	log.Infoln("Generating faucet configuration")

	// Use the first ExtraAccount as a faucet account
	faucetAccount := evt.FaucetAccount()
	if faucetAccount == nil {
		return errors.New("at this stage we expect every blockchain deployment to have a Faucet account")
	}
	// The faucet should connect to one of the validator nodes
	v, _ := evt.Validators()
	nodeIP := evt.State[v[0]].Instance.IPAddress
	out, err := runCommand([]string{"docker", "pull", evt.Payload.DockerImage}, []string{})
	if err != nil {
		return
	}
	log.Debugln(out)

	evtsDir, err := settings.Evts(evt.ID())
	if err != nil {
		return
	}

	// docker image permissions problems again - faucet cannot write to mounted volume
	os.Chmod(filepath.Join(evtsDir, "nodeconfig"), 0777)

	command := []string{"docker", "run", "-v", fmt.Sprintf("%s:/payload/config", filepath.Join(evtsDir, "nodeconfig")), evt.Payload.DockerImage, "/payload/configurefaucet.sh", evt.ID(), faucetAccount.Address, evt.TokenSymbol, nodeIP}
	out, err = runCommand(command, []string{})
	log.Debugln(out)
	return
}

func configurePayload(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	err = DownloadPayloadBinary(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	evt, err = InitDaemon(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = StoreEvent(settings, evt)
	if err != nil {
		return
	}
	evt, err = GenerateKeys(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = StoreEvent(settings, evt)
	if err != nil {
		return
	}
	err = AddGenesisAccounts(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = GenesisTxs(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = CollectGenesisTxs(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = EditConfigs(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	err = GenerateFaucetConfig(settings, evt, cmdRunner)
	if err != nil {
		return
	}
	return
}

// ConfigurePayload is a wrapper function that runs all the needed steps to
// generate a payload's configuration and fills out the evt object with said information.
func ConfigurePayload(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	nodeconfigPath, err := settings.ConfigDir(evt.ID())
	if err != nil {
		return
	}
	err = configurePayload(settings, evt, cmdRunner)
	if err != nil {
		os.RemoveAll(nodeconfigPath)
	}
	return
}
