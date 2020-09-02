package config

type Schema struct {
	Workspace     string        `mapstructure:"workspace,omitempty"`
	DockerMachine DockerMachine `mapstructure:"docker_machine,omitempty"`
	LaunchPayload LaunchPayload `mapstructure:"launch_payload,omitempty"`
}

// DockerMachine holds docker machine configuration
type DockerMachine struct {
	Workspace  string                         `mapstructure:"workspace,omitempty"`
	SearchPath []string                       `mapstructure:"search_path,omitempty"`
	Version    string                         `mapstructure:"version,omitempty"`
	BinaryURL  string                         `mapstructure:"binary_url,omitempty"`
	Binary     string                         `mapstructure:"binary,omitempty"`
	Drivers    map[string]DockerMachineDriver `mapstructure:"drivers,omitempty"`
}

// DockerMachineDriver holds a driver configuration
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
