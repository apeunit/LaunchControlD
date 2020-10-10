package config

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace     string        `mapstructure:"workspace,omitempty"`
	DockerMachine DockerMachine `mapstructure:"docker_machine,omitempty"`
	EventParams   EventParams   `mapstructure:"event_params,omitempty"`
}

// DockerMachine describes the host's docker-machine binary
type DockerMachine struct {
	Workspace  string                         `mapstructure:"workspace,omitempty"`
	SearchPath []string                       `mapstructure:"search_path,omitempty"`
	Version    string                         `mapstructure:"version,omitempty"`
	BinaryURL  string                         `mapstructure:"binary_url,omitempty"`
	Binary     string                         `mapstructure:"binary,omitempty"`
	Drivers    map[string]DockerMachineDriver `mapstructure:"drivers,omitempty"`
}

// DockerMachineDriver describes the location and environment params of any
// optional (non-built-in) docker-machine drivers on the host system
type DockerMachineDriver struct {
	Version   string   `mapstructure:"version,omitempty"`
	BinaryURL string   `mapstructure:"binary_url,omitempty"`
	Binary    string   `mapstructure:"binary,omitempty"`
	Params    []string `mapstructure:"params,omitempty"`
	Env       []string `mapstructure:"env,omitempty"`
}

// LaunchPayload holds metadata about the copy of the launchpayload that is
// stored on the deployer's machine
type LaunchPayload struct {
	BinaryURL  string `mapstructure:"binary_url,omitempty"`
	BinaryPath string `mapstructure:"binary_path,omitempty"`
	DaemonPath string `mapstructure:"daemon_path,omitempty"`
	CLIPath    string `mapstructure:"cli_path,omitempty"`
}

// EventParams specifies how the genesis.json should be setup
type EventParams struct {
	LaunchPayload   LaunchPayload    `mapstructure:"launch_payload,omitempty"`
	DockerImage     string           `mapstructure:"docker_image,omitempty"`
	GenesisAccounts []GenesisAccount `mapstructure:"genesis_accounts,omitempty"`
}

// GenesisAccount is the configuration of accounts present in the genesis block
type GenesisAccount struct {
	Name           string `mapstructure:"name,omitempty"`
	GenesisBalance string `mapstructure:"genesis_balance,omitempty"`
	Validator      bool   `mapstructure:"validator,omitempty"`
}
