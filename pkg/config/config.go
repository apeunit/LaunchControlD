package config

import (
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Folder/file names
const (
	BinDir            = "bin"
	TmpDir            = "tmp"
	EvtsDir           = "evts"
	EvtDescriptorFile = "event.json"
)

// set configuration defaults
func init() {
	// web
	viper.SetDefault("web.listen_address", ":2012")
	viper.SetDefault("web.users_db_file", "users.json")
	viper.SetDefault("web.default_provider", "virtualbox")
	// sentry
	viper.SetDefault("sentry.dsn", "https://17c93719b0a94e139ec731d306648ca1@o413394.ingest.sentry.io/5627329")
	viper.SetDefault("sentry.environment", "develop")
	viper.SetDefault("sentry.enabled", true)
}

// Schema describes the layout of config.yaml
type Schema struct {
	Workspace     string        `mapstructure:"workspace"`
	DockerMachine DockerMachine `mapstructure:"docker_machine"`
	Web           WebSchema     `mapstructure:"web"`
	Sentry        SentrySchema  `mapstructure:"sentry"`
	// the following are used at runtime
	RuntimeStartedAt time.Time `mapstructure:"-"`
	RuntimeVersion   string    `mapstructure:"-"`
}

// DmBin returns /tmp/workspace/bin/docker-machine
func (s *Schema) DmBin() string {
	return s.Bin(s.DockerMachine.Binary)
}

// Bin returns /tmp/workspace/bin/<FILE>
func (s *Schema) Bin(file string) string {
	return filepath.Join(s.Workspace, BinDir, file)
}

// Tmp returns /Tmp/workspace/tmp
func (s *Schema) Tmp() (string, error) {
	return ioutil.TempDir(filepath.Join(s.Workspace, TmpDir), "")
}

// Evts returns /tmp/workspace/evts/<EVTID>
func (s *Schema) Evts(evtID string) (string, error) {
	return filepath.Abs(filepath.Join(s.Workspace, EvtsDir, evtID))
}

// EvtFile returns "/tmp/workspace/evts/<EVTID>/event.json", i.e. the absolute path to the event descriptor file
func (s *Schema) EvtFile(evtID string) (path string, err error) {
	path, err = s.Evts(evtID)
	if err != nil {
		return
	}
	path = filepath.Join(path, EvtDescriptorFile)
	return
}

// ConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfigs
func (s *Schema) ConfigDir(eventID string) (finalPath string, err error) {
	p, err := s.Evts(eventID)
	if err != nil {
		return
	}
	return path.Join(p, "nodeconfig"), nil
}

// NodeConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfig/0
func (s *Schema) NodeConfigDir(eventID, nodeID string) (configDir string, err error) {
	basePath, err := s.ConfigDir(eventID)
	if err != nil {
		return
	}
	nodeIDsplit := strings.Split(nodeID, "-")
	return path.Join(basePath, nodeIDsplit[len(nodeIDsplit)-1]), nil
}

// ExtraAccountConfigDir returns /tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/nodeconfig/extra_accounts
func (s *Schema) ExtraAccountConfigDir(eventID, name string) (finalPath string, err error) {
	p, err := s.ConfigDir(eventID)
	if err != nil {
		return
	}
	return path.Join(p, "extra_accounts", name), nil
}

// SentrySchema configure sentry
type SentrySchema struct {
	Enabled     bool   `mapstructure:"enabled"`
	DSN         string `mapstructure:"dsn"`
	Environment string `mapstructure:"environment"`
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
