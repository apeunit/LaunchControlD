package utils

import (
	"io/ioutil"
	"path/filepath"

	"github.com/apeunit/LaunchControlD/pkg/config"
)

const (
	BinDir            = "bin"
	TmpDir            = "tmp"
	EvtsDir           = "evts"
	EvtDescriptorFile = "event.json"
	DockerHome        = ".docker"
)

// DmBin returns /tmp/workspace/bin/docker-machine
func DmBin(settings config.Schema) string {
	return Bin(settings, settings.DockerMachine.Binary)
}

// Bin returns /tmp/workspace/bin
func Bin(settings config.Schema, file string) string {
	return filepath.Join(settings.Workspace, BinDir, file)
}

// Tmp returns /Tmp/workspace/tmp
func Tmp(settings config.Schema) (string, error) {
	return ioutil.TempDir(filepath.Join(settings.Workspace, TmpDir), "")
}

// Evts returns /tmp/workspace/evts/<EVTID>
func Evts(settings config.Schema, evtID string) (string, error) {
	return filepath.Abs(filepath.Join(settings.Workspace, EvtsDir, evtID))
}

// EvtFile returns "/tmp/workspace/evts/<EVTID>/event.json", i.e. the absolute path to the event descriptor file
func EvtFile(settings config.Schema, evtID string) (path string, err error) {
	path, err = Evts(settings, evtID)
	if err != nil {
		return
	}
	path = filepath.Join(path, EvtDescriptorFile)
	return
}
