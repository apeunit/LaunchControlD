package config

import "github.com/apeunit/LaunchControlD/pkg/model"

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace              string                `mapstructure:"workspace"`
	DockerMachine          DockerMachine         `mapstructure:"docker_machine"`
	DefaultPayloadLocation model.PayloadLocation `mapstructure:"default_payload_location"`
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
