package config

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace     string        `mapstructure:"workspace"`
	DockerMachine DockerMachine `mapstructure:"docker_machine"`
	EventRequest   EventRequest   `mapstructure:"event_params"`
}

// DockerMachine describes the host's docker-machine binary
type DockerMachine struct {
	Workspace  string                         `mapstructure:"workspace"`
	SearchPath []string                       `mapstructure:"search_path"`
	Version    string                         `mapstructure:"version"`
	BinaryURL  string                         `mapstructure:"binary_url"`
	Binary     string                         `mapstructure:"binary"`
	Drivers    map[string]DockerMachineDriver `mapstructure:"drivers"`
}

// DockerMachineDriver describes the location and environment params of any
// optional (non-built-in) docker-machine drivers on the host system
type DockerMachineDriver struct {
	Version   string   `mapstructure:"version"`
	BinaryURL string   `mapstructure:"binary_url"`
	Binary    string   `mapstructure:"binary"`
	Params    []string `mapstructure:"params"`
	Env       []string `mapstructure:"env"`
}

// LaunchPayload holds metadata about the copy of the launchpayload that is
// stored on the deployer's machine
type LaunchPayload struct {
	BinaryURL  string `mapstructure:"binary_url"`
	BinaryPath string `mapstructure:"binary_path"`
	DaemonPath string `mapstructure:"daemon_path"`
	CLIPath    string `mapstructure:"cli_path"`
}

// EventRequest specifies how the genesis.json should be setup
type EventRequest struct {
	LaunchPayload   LaunchPayload    `mapstructure:"launch_payload"`
	DockerImage     string           `mapstructure:"docker_image"`
	GenesisAccounts []GenesisAccount `mapstructure:"genesis_accounts"`
}

// GenesisAccount is the configuration of accounts present in the genesis block
type GenesisAccount struct {
	Name           string `mapstructure:"name"`
	GenesisBalance string `mapstructure:"genesis_balance"`
	Validator      bool   `mapstructure:"validator"`
}
