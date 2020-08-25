package LaunchControlD

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apeunit/lctrld/pkg/config"
	"github.com/apeunit/lctrld/pkg/model"
	"github.com/apeunit/lctrld/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func dockerEnv(settings config.Schema, evt model.EvtvzE) (env []string, err error) {
	// set the path to find executable
	p := append(settings.DockerMachine.SearchPath, bin(settings, ""))
	envPath := fmt.Sprintf("PATH=%s", strings.Join(p, ":"))
	// set the home path for the command
	home, err := evts(settings, evt.ID()) // this gives you the relative path to the event home
	if err != nil {
		return
	}
	envHome := fmt.Sprintf("HOME=%s", home)
	// get the env var from the driver
	env = settings.DockerMachine.Drivers[evt.Provider].Env
	env = append(env, envHome, envPath)
	return
}

// InspectEvent inspect status of the infrastructure for an event
func InspectEvent(settings config.Schema, evt model.EvtvzE) (err error) {
	path, err := evts(settings, evt.ID())
	log.Debugln("InspectEvent event", evt.ID(), "home:", path)
	if err != nil {
		log.Error("Inspect failed:", err)
		return
	}
	dmBin := dmBin(settings)
	// set the path to find the executable
	evnVars, err := dockerEnv(settings, evt)
	for i := range evt.Validators {
		host := evt.NodeID(i)
		out, err := runCommand(dmBin, []string{"status", host}, evnVars)
		if err != nil {
			break
		}
		fmt.Println(host, "status:", out)
		out, err = runCommand(dmBin, []string{"ip", host}, evnVars)
		fmt.Println(host, "IP:", out)
	}
	return
}

// DestroyEvent destroy an existing event
func DestroyEvent(settings config.Schema, evtID string) (err error) {
	path, err := evts(settings, evtID)
	log.Debugln("DestroyEvent event", evtID, "home:", path)
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
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return
	}
	for i, v := range evt.Validators {
		host := evt.NodeID(i)
		//driver := settings.DockerMachine.Drivers[evt.Provider]
		log.Infof("Node ID for %s is %s", v, host)
		// create the parameters
		out, err := runCommand(dmBin, []string{"rm", host}, envVars)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(host, "rm:", out)

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
	evnVars, err := dockerEnv(settings, evt)
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