package model

import (
	"io/ioutil"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"gopkg.in/yaml.v2"
)

// EventRequest holds metadata about the event, including how the genesis.json should be setup
type EventRequest struct {
	TokenSymbol     string           `yaml:"token_symbol" json:"token_symbol"`
	GenesisAccounts []GenesisAccount `yaml:"genesis_accounts" json:"genesis_accounts"`
	PayloadLocation PayloadLocation  `yaml:"payload_location" json:"payload,omitempty"`
	Owner           string           `yaml:"owner" json:"owner,omitempty"`
	Provider        string           `yaml:"provider" json:"provider,omitempty"`
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

// NewDefaultPayloadLocation is a placeholder to help code refactoring and reduce import cycles
func NewDefaultPayloadLocation(settings *config.Schema) PayloadLocation {
	return PayloadLocation{
		DockerImage: "apeunit/launchpayload:v1.0.0",
		BinaryURL:   "https://github.com/apeunit/LaunchPayload/releases/download/v0.0.0/launchpayload-v0.0.0.zip",
		BinaryPath:  settings.Workspace,
		DaemonPath:  settings.Bin("launchpayloadd"),
		CLIPath:     settings.Bin("launchpayloadcli"),
	}
}
