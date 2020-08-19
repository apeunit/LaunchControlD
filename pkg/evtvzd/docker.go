package evtvzd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/apeunit/evtvzd/pkg/config"
	"github.com/apeunit/evtvzd/pkg/model"
	"github.com/apeunit/evtvzd/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func evtEnvVars(settings config.Schema, evt model.EvtvzE) (env []string, err error) {
	// set the path to find the executable
	envPath := fmt.Sprintf("PATH=%s", bin(settings, ""))
	// set the home path for the command
	home, err := evts(settings, evt.ID()) // this gives you the relative path to the event home
	if err != nil {
		return
	}
	envHome := fmt.Sprintf("HOME=%s", home)
	env = []string{envHome, envPath}
	return
}

// DestroyEvent destroy an existing event
func DestroyEvent(settings config.Schema, evtID string) (err error) {
	path, err := evts(settings, evtID)
	log.Debug("DestroyEvent event home:", path)
	if err != nil {
		log.Error("DestroyEvent failed:", err)
		return
	}
	if !utils.FileExists(path) {
		err = fmt.Errorf("Event ID %s not found", evtID)
		log.Error("DestroyEvent failed:", err)
		return
	}
	// load the descriptor
	p, err := evtDescriptor(settings, evtID)
	log.Debug("DestroyEvent event descriptor:", p)
	if err != nil {
		log.Error("DestroyEvent failed:", err)
		return
	}
	evt := model.EvtvzE{}
	err = utils.LoadJSON(p, &evt)
	if err != nil {
		return
	}
	// run the rm command for each validator
	// Outputter
	var out []byte
	dmBin := dmBin(settings)
	// set the path to find the executable
	evnVars, err := evtEnvVars(settings, evt)
	if err != nil {
		return
	}
	for i, v := range evt.Validators {
		host := evt.NodeID(i)
		//driver := settings.DockerMachine.Drivers[evt.Provider]

		log.Infof("Node ID for %s is %s", v, host)
		// create the parameters
		p := []string{"rm"}
		//p = append(p, driver.Params...)
		p = append(p, host)
		log.Debug("DestroyEvent cmd: ", dmBin, host)
		/// prepare the command
		cmd := exec.Command(dmBin, p...)
		// add the binary folder to the exec path
		cmd.Env = evnVars
		log.Debug("DestroyEvent env vars set to ", cmd.Env)
		// execute the command
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("cmd.Run() failed with %s, %s\n", err, out)
			break
		}
		log.Debug("DestroyEvent cmd ouput: ", string(out), err)
	}
	if err != nil {
		return
	}
	err = os.RemoveAll(path)
	return
}

// DeployEvent deploy docker
func DeployEvent(settings config.Schema, evt model.EvtvzE) (err error) {
	// Outputter
	var out []byte
	dmBin := dmBin(settings)
	// set the path to find the executable
	evnVars, err := evtEnvVars(settings, evt)
	if err != nil {
		return
	}

	for i, v := range evt.Validators {
		host := evt.NodeID(i)
		driver := settings.DockerMachine.Drivers[evt.Provider]

		log.Infof("Node ID for %s is %s", v, host)
		// create the parameters
		p := []string{"create", "--driver", evt.Provider}
		p = append(p, driver.Params...)
		p = append(p, host)
		log.Debug("DeployEvent cmd: ", dmBin, evt.Provider, host)
		/// prepare the command
		cmd := exec.Command(dmBin, p...)
		// add the binary folder to the exec path
		cmd.Env = evnVars
		log.Debug("DeployEvent env vars set to ", cmd.Env)
		// execute the command
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("cmd.Run() failed with %s, %s\n", err, out)
			break
		}
		log.Debug("DeployEvent cmd ouput: ", string(out), err)
	}
	if err != nil {
		return
	}
	p, err := evtDescriptor(settings, evt.ID())
	if err != nil {
		return
	}
	err = utils.StoreJSON(p, evt)
	return
}