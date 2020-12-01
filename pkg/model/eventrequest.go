package model

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// EventRequest holds metadata about the event, including how the genesis.json should be setup
type EventRequest struct {
	Payload         Payload          `yaml:"payload" json:"payload"`
	Owner           string           `yaml:"owner" json:"owner"`
	TokenSymbol     string           `yaml:"token_symbol" json:"token_symbol"`
	GenesisAccounts []GenesisAccount `yaml:"genesis_accounts" json:"genesis_accounts"`
}

// Payload holds metadata about the copy of the launchpayload that is
// stored on the machine running LaunchControlD
type Payload struct {
	DockerImage string `yaml:"docker_image" json:"docker_image"`
	BinaryURL   string `yaml:"binary_url" json:"binary_url"`
	BinaryPath  string `yaml:"binary_path" json:"binary_path"`
	DaemonPath  string `yaml:"daemon_path" json:"daemon_path"`
	CLIPath     string `yaml:"cli_path" json:"cli_path"`
}

// GenesisAccount is the configuration of accounts present in the genesis block
type GenesisAccount struct {
	Name           string `yaml:"name" json:"name"`
	GenesisBalance string `yaml:"genesis_balance" json:"genesis_balance"`
	Validator      bool   `yaml:"validator" json:"validator"`
	Faucet         bool   `yaml:"faucet" json:"faucet"`
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
