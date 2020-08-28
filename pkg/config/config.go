package config

type Schema struct {
	Workspace     string        `mapstructure:"workspace,omitempty"`
	DockerMachine DockerMachine `mapstructure:"docker_machine,omitempty"`
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

// LaunchPayload holds metadata about launchpayloadd/launchpayloadcli
type LaunchPayload struct {
	BinaryURL string `mapstructure:"binary_url,omitempty"`
	Binary    string `mapstructure:"binary,omitempty"`
}
