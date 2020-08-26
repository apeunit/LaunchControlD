package lctrld

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"
)

//SetupWorkspace setup the workspace for the service
func SetupWorkspace(settings config.Schema) (err error) {
	// workspace folder
	if !utils.FileExists(settings.Workspace) {
		log.Debugln("Folder ", settings.Workspace, "does not exists, creating")
		err = os.MkdirAll(settings.Workspace, 0700)
		if err != nil {
			log.Error("SetupWorkspace: ", err)
			return
		}
	}

	for _, dirN := range []string{"bin", "tmp", "evts"} {
		dir := filepath.Join(settings.Workspace, dirN)
		if !utils.FileExists(dir) {
			log.Debugln("Folder", dir, "does not exists, creating")
			err = os.MkdirAll(dir, 0700)
			if err != nil {
				log.Error("SetupWorkspace: ", err)
				return
			}
		}
	}
	return
}

// InstallDockerMachine setup docker machine environment
func InstallDockerMachine(settings config.Schema) (err error) {
	log.Debug("InstallDockerMachine setup binaries")

	// download if not exists helper
	dine := func(file, downloadURL string) (err error) {
		targetPath := bin(settings, file)
		log.Debug("InstallDockerMachine: checking ", targetPath)
		if utils.FileExists(targetPath) {
			log.Debug("InstallDockerMachine: ", targetPath, " found!")
			return
		}
		log.Debug("InstallDockerMachine: ", targetPath, " does not exists, downloading from ", downloadURL)
		// generate a temp dir
		td, err := tmp(settings)
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
			return
		}
		log.Debug("InstallDockerMachine: file will be download in ", td)
		dwnFile, err := utils.DownloadFile(td, downloadURL)
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
			return
		}
		dwnFilePath := _path(td, dwnFile)
		log.Debug("InstallDockerMachine: download complete ", dwnFilePath)
		ct, err := utils.DetectContentType(dwnFilePath)
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
			return
		}
		log.Debug("InstallDockerMachine: downloaded file ", dwnFilePath, " has content-type ", ct)
		switch ct {
		case "application/octet-stream":
			log.Debugln("InstallDockerMachine: downloaded file is binary, moving to the destination path")
			err = os.Rename(dwnFilePath, targetPath)
			if err != nil {
				log.Error("InstallDockerMachine: ", err)
				return
			}
		case "application/zip":
			err = fmt.Errorf("InstallDockerMachine: unsupported file type %s", ct)
		case "application/x-gzip":
			err = utils.ExtractGzip(dwnFilePath, td)
			if err != nil {
				log.Error("InstallDockerMachine: ", err)
				return
			}
			err = utils.SearchAndMove(td, file, targetPath)
			if err != nil {
				log.Error("InstallDockerMachine: ", err)
				return
			}
		default:
			err = fmt.Errorf("InstallDockerMachine: unsupported file type %s", ct)
		}
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
			return
		}
		// make it executable
		os.Chmod(targetPath, 0700)
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
		}
		return
	}

	// check if the system has been setup already
	err = dine(settings.DockerMachine.Binary, settings.DockerMachine.BinaryURL)
	if err != nil {
		log.Error("InstallDockerMachine: ", err)
		return
	}
	for dName, driver := range settings.DockerMachine.Drivers {
		log.Debugln("InstallDockerMachine: processing driver", dName)
		if len(driver.BinaryURL) == 0 {
			log.Debugln("InstallDockerMachine: driver", dName, "does not require installation (download url not provided)")
			continue
		}
		err = dine(driver.Binary, driver.BinaryURL)
		if err != nil {
			log.Error("InstallDockerMachine: ", err)
			return
		}
	}
	return
}

// CreateEvent creates the event home and the event descriptor
func CreateEvent(settings config.Schema, evt model.EvtvzE) (err error) {
	path, err := evts(settings, evt.ID())
	if !utils.FileExists(path) {
		err = os.MkdirAll(path, 0700)
		if err != nil {
			return
		}
	}
	err = storeEvent(settings, evt)
	return
}

// ListEvents list available events
func ListEvents(settings config.Schema) (events []model.EvtvzE, err error) {
	evtsBase, err := evts(settings, "")
	if err != nil {
		log.Error("ListEvents failed:", err)
		return
	}
	filepath.Walk(evtsBase, func(subPath string, info os.FileInfo, err error) error {
		if info.Name() == dockerHome {
			log.Debugln("Folder", info.Name(), "skipped")
			// skip docker folder
			return filepath.SkipDir
		}
		if info.Name() == evtDescriptorFile {
			log.Debugln("Event found", info.Name())
			evt := model.EvtvzE{}
			err := utils.LoadJSON(subPath, &evt)
			if err != nil {
				log.Error("ListEvents failed:", err)
				return err
			}
			events = append(events, evt)
		}
		return nil
	})
	return
}
