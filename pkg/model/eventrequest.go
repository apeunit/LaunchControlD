package model

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// EventRequest specifies how the genesis.json should be setup
type EventRequest struct {
	Payload         Payload          `yaml:"payload"`
	Owner           string           `yaml:"owner"`
	TokenSymbol     string           `yaml:"token_symbol"`
	GenesisAccounts []GenesisAccount `yaml:"genesis_accounts"`
}

// Payload holds metadata about the copy of the launchpayload that is
// stored on the machine running LaunchControlD
type Payload struct {
	DockerImage string `yaml:"docker_image"`
	BinaryURL   string `yaml:"binary_url"`
	BinaryPath  string `yaml:"binary_path"`
	DaemonPath  string `yaml:"daemon_path"`
	CLIPath     string `yaml:"cli_path"`
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
