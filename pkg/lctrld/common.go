package lctrld

import (
	"io/ioutil"
	"path/filepath"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
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
func evts(settings config.Schema, evtID string) (string, error) {
	return _absPath(_path(settings.Workspace, evtsDir, evtID))
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
