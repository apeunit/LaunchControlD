package lctrld

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	binDir            = "bin"
	tmpDir            = "tmp"
	evtsDir           = "evts"
	evtDescriptorFile = "event.json"
	dockerHome        = ".docker"
)

func bin(settings config.Schema, file string) string {
	return _path(settings.Workspace, binDir, file)
}

func tmp(settings config.Schema) (string, error) {
	return ioutil.TempDir(_path(settings.Workspace, tmpDir), "")
}

func evts(settings config.Schema, dir string) (string, error) {
	return _absPath(_path(settings.Workspace, evtsDir, dir))
}

type DockerMachineInterface interface {
	HomeDir(int) string
	ReadConfig(int) (*model.MachineConfig, error)
}

// DockerMachineConfig holds information that lets lctrld read the state of a
// docker-machine provisioning
type DockerMachineConfig struct {
	EventID  string
	Settings config.Schema
}

// NewDockerMachineConfig ensures that all fields of a DockerMachineConfig are filled out
func NewDockerMachineConfig(settings config.Schema, eventID string) *DockerMachineConfig {
	return &DockerMachineConfig{
		EventID:  eventID,
		Settings: settings,
	}
}

// HomeDir get the path of a docker-machine instance home
func (dmc *DockerMachineConfig) HomeDir(machineN int) string {
	return _path(dmc.Settings.Workspace, evtsDir, dmc.EventID, ".docker", "machine", "machines", fmt.Sprintf("%s-%d", dmc.EventID, machineN))

}

// ReadConfig return configuration of a docker machine
func (dmc *DockerMachineConfig) ReadConfig(machineN int) (mc model.MachineConfig, err error) {
	err = utils.LoadJSON(_path(dmc.HomeDir(machineN), "config.json"), &mc)
	return
}

// evtDescriptor returns the absolute path to the event descriptor file
func evtDescriptor(settings config.Schema, evtID string) (path string, err error) {
	path, err = evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	return
}

//LoadEvent returns the Event model of the specified event ID
func LoadEvent(settings config.Schema, evtID string) (evt *model.EvtvzE, err error) {
	path, err := evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.LoadJSON(path, &evt)
	return
}

// StoreEvent returns the Event model of the specified event ID
func StoreEvent(settings config.Schema, evt *model.EvtvzE) (err error) {
	path, err := evts(settings, evt.ID())
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.StoreJSON(path, evt)
	return
}

func _path(pieces ...string) string {
	return filepath.Join(pieces...)
}

func _absPath(relative string) (string, error) {
	return filepath.Abs(relative)
}

func dmBin(settings config.Schema) string {
	return bin(settings, settings.DockerMachine.Binary)
}

func dmDriverBin(settings config.Schema, driverName string) string {
	return bin(settings, settings.DockerMachine.Drivers[driverName].Binary)
}

type commandRunner interface {
	Run(string, []string, []string) (string, error)
}
type CommandRunner struct{}

// NewCommandRunner returns a pointer to a new CommandRunner instance
func NewCommandRunner() *CommandRunner {
	return new(CommandRunner)
}

// Run is a pure function implemented as a method so that I can use it in an
// interface for mocking
func (c *CommandRunner) Run(bin string, args, envVars []string) (out string, err error) {
	/// prepare the command
	cmd := exec.Command(bin, args...)
	// add the binary folder to the exec path
	cmd.Env = envVars
	log.Debug("command env vars set to ", cmd.Env)
	// execute the command
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("%s %s failed with %s, %s\n", bin, args, err, string(o))
		return
	}
	out = strings.TrimSpace(string(o))
	log.Debug("command stdout: ", out)
	return
}
