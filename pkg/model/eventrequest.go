package model

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// EventRequest holds metadata about the event, including how the genesis.json should be setup
type EventRequest struct {
	PayloadLocation PayloadLocation  `yaml:"payload_location" json:"payload"`
	Owner           string           `yaml:"owner" json:"owner"`
	TokenSymbol     string           `yaml:"token_symbol" json:"token_symbol"`
	GenesisAccounts []GenesisAccount `yaml:"genesis_accounts" json:"genesis_accounts"`
	Provider        string           `yaml:"provider" json:"provider"`
}

// PayloadLocation holds metadata about the copy of the launchpayload that is
// stored on the machine running LaunchControlD
type PayloadLocation struct {
	DockerImage string `mapstructure:"docker_image" yaml:"docker_image" json:"docker_image"`
	BinaryURL   string `mapstructure:"binary_url" yaml:"binary_url" json:"binary_url"`
	BinaryPath  string `mapstructure:"binary_path" yaml:"binary_path" json:"binary_path"`
	DaemonPath  string `mapstructure:"daemon_path" yaml:"daemon_path" json:"daemon_path"`
	CLIPath     string `mapstructure:"cli_path" yaml:"cli_path" json:"cli_path"`
}

// GenesisAccount is the configuration of accounts present in the genesis block
type GenesisAccount struct {
	Name           string `yaml:"name" json:"name"`
	GenesisBalance string `yaml:"genesis_balance" json:"genesis_balance"`
	Validator      bool   `yaml:"validator" json:"validator"`
	Faucet         bool   `yaml:"faucet" json:"faucet"`
}

// ParseEventRequest from a json object
func ParseEventRequest(jsonData []byte, pl PayloadLocation, provider string) (er EventRequest, err error) {
	err = json.Unmarshal(jsonData, &er)
	if err != nil {
		return
	}
	// set the default PayloadLocation
	er.PayloadLocation = pl
	// set the default provider
	er.Provider = provider
	return
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
