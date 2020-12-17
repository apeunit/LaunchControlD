package config

import (
	"time"

	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/spf13/viper"
)

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace              string                `mapstructure:"workspace"`
	DockerMachine          DockerMachine         `mapstructure:"docker_machine"`
	DefaultPayloadLocation model.PayloadLocation `mapstructure:"default_payload_location"`
	Web                    WebSchema             `mapstructure:"web"`
	// the following are used at runtime
	RuntimeStartedAt time.Time `mapstructure:"-"`
	RuntimeVersion   string    `mapstructure:"-"`
}

// WebSchema configuration for web
type WebSchema struct {
	ListenAddress   string `mapstructure:"listen_address"`
	DefaultProvider string `mapstructure:"default_provider"`
	UsersDbFile     string `mapstructure:"users_db_file"`
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

// Defaults configure defaults for the configuration
func Defaults() {
	// web
	viper.SetDefault("web.listen_address", ":2012")
	viper.SetDefault("web.users_db_file", "users.json")
	viper.SetDefault("web.default_provider", "virtualbox")
	// services
	// viper.SetDefault("services.sentry_dsn", "")
}
