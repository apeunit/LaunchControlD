package model

// Account represents an Account for a Event
type Account struct {
	Name           string         `json:"name"`
	Address        string         `json:"address"`
	Mnemonic       string         `json:"mnemonic"`
	GenesisBalance string         `json:"genesis_balance"`
	Validator      bool           `json:"validator"`
	Faucet         bool           `json:"faucet"`
	ConfigLocation ConfigLocation `json:"config_location"`
}

// ConfigLocation holds the paths to the configuration files for the Cosmos-SDK
// based node and CLI.
type ConfigLocation struct {
	CLIConfigDir    string `json:"CLIConfigDir"`
	DaemonConfigDir string `json:"DaemonConfigDir"`
}
