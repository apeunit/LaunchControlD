package config

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

const (
	BinDir            = "bin"
	TmpDir            = "tmp"
	EvtsDir           = "evts"
	EvtDescriptorFile = "event.json"
	DockerHome        = ".docker"
)

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace     string        `mapstructure:"workspace"`
	DockerMachine DockerMachine `mapstructure:"docker_machine"`
	Web           WebSchema     `mapstructure:"web"`
	// the following are used at runtime
	RuntimeStartedAt time.Time `mapstructure:"-"`
	RuntimeVersion   string    `mapstructure:"-"`
}

// DmBin returns /tmp/workspace/bin/docker-machine
func (s Schema) DmBin() string {
	return s.Bin(s.DockerMachine.Binary)
}

// Bin returns /tmp/workspace/bin/<FILE>
func (s Schema) Bin(file string) string {
	return filepath.Join(s.Workspace, BinDir, file)
}

// Tmp returns /Tmp/workspace/tmp
func (s Schema) Tmp() (string, error) {
	return ioutil.TempDir(filepath.Join(s.Workspace, TmpDir), "")
}

// Evts returns /tmp/workspace/evts/<EVTID>
func (s Schema) Evts(evtID string) (string, error) {
	return filepath.Abs(filepath.Join(s.Workspace, EvtsDir, evtID))
}

// EvtFile returns "/tmp/workspace/evts/<EVTID>/event.json", i.e. the absolute path to the event descriptor file
func (s Schema) EvtFile(evtID string) (path string, err error) {
	path, err = s.Evts(evtID)
	if err != nil {
		return
	}
	path = filepath.Join(path, EvtDescriptorFile)
	return
}

// WebSchema configuration for web
type WebSchema struct {
	ListenAddress   string `mapstructure:"listen_address"`
	DefaultProvider string `mapstructure:"default_provider"`
	UsersDbFile     string `mapstructure:"users_db_file"`
}

// DockerMachine describes the host's docker-machine binary
type DockerMachine struct {
	Version   string                         `mapstructure:"version"`
	BinaryURL string                         `mapstructure:"binary_url"`
	Binary    string                         `mapstructure:"binary"`
	Drivers   map[string]DockerMachineDriver `mapstructure:"drivers"`
	Env       []string                       `mapstructure:"env"`
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
