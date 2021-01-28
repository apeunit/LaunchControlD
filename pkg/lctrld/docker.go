package lctrld

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"
)

// InspectEvent inspect status of the infrastructure for an event
func InspectEvent(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	path, err := settings.Evts(evt.ID())
	log.Debugln("InspectEvent event", evt.ID(), "home:", path)
	if err != nil {
		log.Error("Inspect failed:", err)
		return
	}
	dm := NewDockerMachine(settings, evt.ID())
	_, validatorAccounts := evt.Validators()
	for i := range validatorAccounts {
		machineName := evt.NodeID(i)
		out, err := dm.Status(machineName, cmdRunner)
		if err != nil {
			break
		}
		fmt.Println(machineName, out)
	}
	return
}

// DestroyEvent destroy an existing event
func DestroyEvent(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	path, err := settings.Evts(evt.ID())
	log.Debugln("op DestroyEvent event", evt.ID(), "home:", path)
	if err != nil {
		// TODO: this is going to crash the program
		log.Error("DestroyEvent failed:", err)
		return
	}
	if !utils.FileExists(path) {
		err = fmt.Errorf("event ID %s not found", evt.ID())
		log.Error("op DestroyEvent failed:", err)
		return
	}
	// load the descriptor
	p, err := settings.EvtFile(evt.ID())
	log.Debug("op DestroyEvent event descriptor:", p)
	if err != nil {
		// TODO: this is going to crash the program
		log.Error("op DestroyEvent failed:", err)
		return
	}

	dm := NewDockerMachine(settings, evt.ID())
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		machineName := evt.NodeID(i)
		log.Infof("%s's node ID is %s", v.Name, machineName)
		err = dm.StopMachine(machineName, cmdRunner)
		if err != nil {
			return
		}
	}
	err = os.RemoveAll(path)
	return
}

// ProvisionEvent provision the infrastructure for the event
func ProvisionEvent(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	dm := NewDockerMachine(settings, evt.ID())
	// init docker nodes map
	// TODO: shouldn't this be initialized already during evt struct creation?
	evt.State = make(map[string]*model.Machine)
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		machineName := evt.NodeID(i)

		log.Infof("%s's node ID is %s", v.Name, machineName)
		// create the parameters
		mc, err2 := dm.ProvisionMachine(machineName, evt.Provider, cmdRunner)
		if err2 != nil {
			return err2
		}
		evt.State[v.Name] = mc
	}
	log.Infof("Your event ID is %s", evt.ID())
	return
}

// RereadDockerMachineInfo is useful when docker-machine failed during 'create',
// and a human fixed the problem, and wants to continue
func RereadDockerMachineInfo(settings *config.Schema, evt *model.Event) (event *model.Event, err error) {
	dm := NewDockerMachine(settings, evt.ID())
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		machineName := fmt.Sprintf("%s-%s", evt.ID(), i)
		mc, err := dm.ReadConfig(machineName)
		if err != nil {
			log.Error("Provision read machine config error:", err)
			return nil, err
		}
		evt.State[v.Name] = mc
	}
	return evt, err
}

// DeployPayload tells the provisioned machines to run the configured docker image
func DeployPayload(settings *config.Schema, evt *model.Event, cmdRunner cmdrunner.CommandRunner) (err error) {
	var command []string
	log.Infoln("Copying node configs to each provisioned machine")

	dm := NewDockerMachine(settings, evt.ID())
	for name, state := range evt.State {
		// docker-machine ssh mkdir -p /home/docker/nodeconfig
		command = []string{"mkdir", "-p", "/home/docker/nodeconfig"}
		_, err = dm.Run(state.ID(), command, cmdRunner)
		if err != nil {
			return
		}

		// docker-machine scp -r pathDaemon evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		err = dm.Copy(state.ID(), evt.Accounts[name].ConfigLocation.DaemonConfigDir, "/home/docker/nodeconfig", cmdRunner)
		if err != nil {
			return
		}

		// docker-machine scp -r pathCLI evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		err = dm.Copy(state.ID(), evt.Accounts[name].ConfigLocation.CLIConfigDir, "/home/docker/nodeconfig", cmdRunner)
		if err != nil {
			return
		}

		// docker-machine chmod -R 777 /home/docker/nodeconfig
		command = []string{"chmod", "-R", "777", "/home/docker/nodeconfig"}
		_, err = dm.Run(state.ID(), command, cmdRunner)
		if err != nil {
			return
		}
	}

	log.Infof("Running docker pull %s on each provisioned machine", evt.Payload.DockerImage)
	for email, state := range evt.State {
		// in docker-machine provisioned machine: docker pull apeunit/launchpayload
		command := []string{"pull", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine\n", command, email)
		_, err = dm.RunDocker(state.ID(), command, cmdRunner)
		if err != nil {
			return
		}
	}

	log.Infoln("Running the dockerized Cosmos daemons on the provisioned machines")
	for email, state := range evt.State {
		// in docker-machine provisioned machine: docker run -v /home/docker/nodeconfig:/payload/config apeunit/launchpayload
		command := []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "26656:26656", "-p", "26657:26657", "-p", "26658:26658", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine\n", command, email)
		_, err = dm.RunDocker(state.ID(), command, cmdRunner)
		if err != nil {
			return
		}
	}

	// https://forum.cosmos.network/t/what-could-cause-sync-mutex-lock-to-have-a-nil-pointer-dereference/4194
	log.Infoln("Running the CLI to provide the Light Client Daemon")
	v, _ := evt.Validators()
	firstValidator := evt.State[v[0]]
	command = []string{"run", "-d", "--volume=/home/docker/nodeconfig:/payload/config", "-p", "1317:1317", evt.Payload.DockerImage, "/payload/runlightclient.sh", firstValidator.Instance.IPAddress, evt.ID()}
	// command = []string{"scp", evt.Payload.CLIPath, fmt.Sprintf("%s:/home/docker", evt.State[firstValidator].ID())}
	log.Debugf("Running docker-machine %s on validator %s machine\n", command, firstValidator)
	_, err = dm.RunDocker(firstValidator.ID(), command, cmdRunner)
	if err != nil {
		return
	}

	log.Infoln("Copying the faucet account and configuration to the first validator machine")
	faucetAccount := evt.FaucetAccount()
	err = dm.Copy(firstValidator.ID(), faucetAccount.ConfigLocation.CLIConfigDir, "/home/docker/nodeconfig/faucet_account", cmdRunner)
	if err != nil {
		return
	}
	// docker-machine chmod -R 777 /home/docker/nodeconfig AGAIN - what a mess!
	command = []string{"chmod", "-R", "777", "/home/docker/nodeconfig"}
	_, err = dm.Run(firstValidator.ID(), command, cmdRunner)
	if err != nil {
		return
	}

	evtDir, err := settings.Evts(evt.ID())
	if err != nil {
		log.Error(err)
		return
	}

	err = dm.Copy(firstValidator.ID(), filepath.Join(evtDir, "nodeconfig", "faucetconfig.yml"), "/home/docker/nodeconfig", cmdRunner)
	if err != nil {
		return
	}

	log.Infoln("Starting the faucet")
	command = []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "8000:8000", evt.Payload.DockerImage, "/payload/runfaucet.sh"}
	log.Debugf("Running docker %s on %s\n", command, firstValidator.ID())
	_, err = dm.RunDocker(firstValidator.ID(), command, cmdRunner)
	return
}
