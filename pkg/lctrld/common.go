package lctrld

import (
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

func _path(pieces ...string) string {
	return filepath.Join(pieces...)
}

func _absPath(relative string) (string, error) {
	return filepath.Abs(relative)
}

// dmBin returns /tmp/workspace/bin/docker-machine
func dmBin(settings config.Schema) string {
	return bin(settings, settings.DockerMachine.Binary)
}

// bin returns /tmp/workspace/bin
func bin(settings config.Schema, file string) string {
	return _path(settings.Workspace, binDir, file)
}

// tmp returns /tmp/workspace/tmp
func tmp(settings config.Schema) (string, error) {
	return ioutil.TempDir(_path(settings.Workspace, tmpDir), "")
}

// evts returns /tmp/workspace/evts/<EVTID>
func evts(settings config.Schema, dir string) (string, error) {
	return _absPath(_path(settings.Workspace, evtsDir, dir))
}

// evtFile returns "/tmp/workspace/evts/<EVTID>/event.json", i.e. the absolute path to the event descriptor file
func evtFile(settings config.Schema, evtID string) (path string, err error) {
	path, err = evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	return
}

//LoadEvent returns the Event model of the specified event ID
func LoadEvent(settings config.Schema, evtID string) (evt *model.Event, err error) {
	path, err := evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.LoadJSON(path, &evt)
	return
}

// StoreEvent saves the Event model to a file
func StoreEvent(settings config.Schema, evt *model.Event) (err error) {
	path, err := evts(settings, evt.ID())
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.StoreJSON(path, evt)
	return
}

// CommandRunner func type allows for mocking out RunCommand()
type CommandRunner func([]string, []string) (string, error)

// RunCommand runs a command
func RunCommand(command, envVars []string) (out string, err error) {
	bin := command[0]
	args := command[1:]

	/// prepare the command
	cmd := exec.Command(bin, args...)
	// add the binary folder to the exec path
	cmd.Env = envVars
	log.Debug("command env vars set to ", cmd.Env)
	// execute the command
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("%s %s failed with %s, %s\n", bin, args, err, string(o))
		return
	}
	out = strings.TrimSpace(string(o))
	log.Debug("command stdout: ", out)
	return
}
