package model

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// EventRequest specifies how the genesis.json should be setup
type EventRequest struct {
	LaunchPayload   LaunchPayload    `yaml:"launch_payload"`
	DockerImage     string           `yaml:"docker_image"`
	Owner           string           `yaml:"owner"`
	TokenSymbol     string           `yaml:"token_symbol"`
	GenesisAccounts []GenesisAccount `yaml:"genesis_accounts"`
}

// LaunchPayload holds metadata about the copy of the launchpayload that is
// stored on the deployer's machine
type LaunchPayload struct {
	BinaryURL  string `yaml:"binary_url"`
	BinaryPath string `yaml:"binary_path"`
	DaemonPath string `yaml:"daemon_path"`
	CLIPath    string `yaml:"cli_path"`
}

// GenesisAccount is the configuration of accounts present in the genesis block
type GenesisAccount struct {
	Name           string `yaml:"name"`
	GenesisBalance string `yaml:"genesis_balance"`
	Validator      bool   `yaml:"validator"`
}

// LoadEventRequestFromFile is as convenience function to unmarshal a EventRequest from a YAML file
func LoadEventRequestFromFile(path string) (eq *EventRequest, err error) {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(f, &eq)
	if err != nil {
		return
	}
	return eq, nil
}
