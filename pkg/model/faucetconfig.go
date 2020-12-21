package model

import "gopkg.in/yaml.v2"

// FaucetConfig describes the configuration file used by the faucet. TODO:
// lctrld shouldn't know so much about the faucet
type FaucetConfig struct {
	ListenAddr    string `yaml:"listen_addr"`
	ChainID       string `yaml:"chain_id"`
	CliBinaryPath string `yaml:"cli_binary_path"`
	CliConfigPath string `yaml:"cli_config_path"`
	FaucetAddr    string `yaml:"faucet_addr"`
	Unit          string `yaml:"unit"`
	NodeAddr      string `yaml:"node_addr"`
	Secret        string `yaml:"secret"`
}

// Parse takes a bytestream and parses it as YAML
func (fc *FaucetConfig) Parse(data []byte) error {
	return yaml.Unmarshal(data, fc)
}
