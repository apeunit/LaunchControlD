package evtvzd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/apeunit/evtvzd/pkg/config"
)

func bin(settings config.Schema, file string) string {
	return _path(settings.Workspace, "bin", file)
}

func tmp(settings config.Schema) (string, error) {
	return ioutil.TempDir(_path(settings.Workspace, "tmp"), "")
}

func evts(settings config.Schema, dir string) (string, error) {
	return _absPath(_path(settings.Workspace, "evts", dir))
}

func evtDescriptor(settings config.Schema, evtID string) (path string, err error) {
	path, err = evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, "event.json")
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
