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

// InspectEvent inspect status of the infrastructure for an event
func InspectEvent(settings config.Schema, evt *model.Event, cmdRunner CommandRunner) (err error) {
	path, err := evts(settings, evt.ID())
	log.Debugln("InspectEvent event", evt.ID(), "home:", path)
	if err != nil {
		log.Error("Inspect failed:", err)
		return
	}
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerMachineEnv(settings, evt)
	_, validatorAccounts := evt.Validators()
	for i := range validatorAccounts {
		host := evt.NodeID(i)
		out, err := cmdRunner([]string{dmBin, "status", host}, envVars)
		if err != nil {
			break
		}
		fmt.Println(host, "status:", out)
		out, err = cmdRunner([]string{dmBin, "ip", host}, envVars)
		if err != nil {
			break
		}
		fmt.Println(host, "IP:", out)
	}
	return
}

// DestroyEvent destroy an existing event
func DestroyEvent(settings config.Schema, evt *model.Event, cmdRunner CommandRunner) (err error) {
	path, err := evts(settings, evt.ID())
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
	p, err := evtFile(settings, evt.ID())
	log.Debug("op DestroyEvent event descriptor:", p)
	if err != nil {
		// TODO: this is going to crash the program
		log.Error("op DestroyEvent failed:", err)
		return
	}

	// run the rm command for each validator
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerMachineEnv(settings, evt)
	if err != nil {
		return
	}

	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		host := evt.NodeID(i)
		//driver := settings.DockerMachine.Drivers[evt.Provider]
		log.Infof("%s's node ID is %s", v.Name, host)
		// create the parameters
		out, err := cmdRunner([]string{dmBin, "stop", host}, envVars)
		if err != nil {
			// TODO: preferr logging over printing to stdout
			fmt.Println(err)
		}
		// TODO: prefer logging over printing on stdout
		fmt.Println(host, "stop:", out)

		out, err = cmdRunner([]string{dmBin, "rm", host}, envVars)
		if err != nil {
			// TODO: prefer logging over printing to stdout
			fmt.Println(err)
		}
		// TODO: preferr logging over printing to stdout
		fmt.Println(host, "rm:", out)

	}
	if err != nil {
		return
	}
	err = os.RemoveAll(path)
	return
}

// Provision provision the infrastructure for the event
func Provision(settings config.Schema, evt *model.Event, cmdRunner CommandRunner, dmc DockerMachineInterface) (err error) {
	// Outputter
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerMachineEnv(settings, evt)
	if err != nil {
		return err
	}
	// init docker nodes map
	evt.State = make(map[string]*model.MachineConfig)
	// run the thing
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		host := evt.NodeID(i)
		driver := settings.DockerMachine.Drivers[evt.Provider]

		log.Infof("%s's node ID is %s", v.Name, host)
		// create the parameters
		p := []string{dmBin, "--debug", "create", "--driver", evt.Provider, "--engine-install-url", "https://releases.rancher.com/install-docker/19.03.9.sh"}
		p = append(p, driver.Params...)
		p = append(p, host)

		log.Debugf("Provision cmd: %s", p)
		log.Debug("Provision env vars set to ", envVars)
		out, errI := cmdRunner(p, envVars)
		if errI != nil {
			err = fmt.Errorf("Provision cmd failed with %s, %s", err, out)
			log.Error(err)
			break
		}

		log.Debug("Provision cmd output: ", string(out), err)
		// load the configuration of the machine
		mc, errI := dmc.ReadConfig(fmt.Sprint(i))
		if errI != nil {
			err = fmt.Errorf("Provision read machine config error: %v", err)
			log.Error(err)
			break
		}
		evt.State[v.Name] = mc
	}
	if err != nil {
		return err
	}
	log.Infof("Your event ID is %s", evt.ID())
	return
}

// RereadDockerMachineInfo is useful when docker-machine failed during 'create',
// and a human fixed the problem, and wants to continue
func RereadDockerMachineInfo(settings config.Schema, evt *model.Event, dmc DockerMachineInterface) (event *model.Event, err error) {
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		mc, err := dmc.ReadConfig(fmt.Sprint(i))
		if err != nil {
			log.Error("Provision read machine config error:", err)
			return nil, err
		}
		evt.State[v.Name] = mc
	}
	return evt, err
}

// DeployPayload tells the provisioned machines to run the configured docker image
func DeployPayload(settings config.Schema, evt *model.Event, cmdRunner CommandRunner, dmc DockerMachineInterface) (err error) {
	dmBin := dmBin(settings)
	var args []string

	log.Infoln("Copying node configs to each provisioned machine")
	for name, state := range evt.State {
		envVars, err := dockerMachineEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerMachineEnv() failed while generating envVars: %s", err)
			return err
		}

		// docker-machine ssh mkdir -p /home/docker/nodeconfig
		command := []string{dmBin, "ssh", state.ID(), "mkdir", "-p", "/home/docker/nodeconfig"}
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", command, err)
			return err
		}

		// docker-machine scp -r pathDaemon evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		command = []string{dmBin, "scp", "-r", evt.Accounts[name].ConfigLocation.DaemonConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID())}
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", command, err)
			return err
		}

		// docker-machine scp -r pathCLI evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		command = []string{dmBin, "scp", "-r", evt.Accounts[name].ConfigLocation.CLIConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID())}
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", command, err)
			return err
		}

		// docker-machine chmod -R 777 /home/docker/nodeconfig
		command = []string{dmBin, "ssh", state.ID(), "chmod", "-R", "777", "/home/docker/nodeconfig"}
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", command, err)
			return err
		}
	}

	log.Infof("Running docker pull %s on each provisioned machine", evt.Payload.DockerImage)
	for email, state := range evt.State {
		envVars, err := dockerMachineEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerMachineEnv() failed while generating envVars: %s", err)
			return err
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(state.N), state)

		// in docker-machine provisioned machine: docker pull apeunit/launchpayload
		command := []string{"docker", "pull", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", command, email, envVars)
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker %s failed with %s", command, err)
			return err
		}
	}

	log.Infoln("Running the dockerized Cosmos daemons on the provisioned machines")
	for email, state := range evt.State {
		envVars, err := dockerMachineEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerMachineEnv() failed while generating envVars: %s", err)
			return err
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(state.N), state)

		// in docker-machine provisioned machine: docker run -v /home/docker/nodeconfig:/payload/config apeunit/launchpayload
		command := []string{"docker", "run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "26656:26656", "-p", "26657:26657", "-p", "26658:26658", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", command, email, envVars)
		_, err = cmdRunner(command, envVars)
		if err != nil {
			log.Errorf("docker %s failed with %s", command, err)
			return err
		}
	}

	// https://forum.cosmos.network/t/what-could-cause-sync-mutex-lock-to-have-a-nil-pointer-dereference/4194
	log.Infoln("Running the CLI to provide the Light Client Daemon")
	emails, _ := evt.Validators()
	firstNode := emails[0]
	machineHomeDir := dmc.HomeDir(evt.State[firstNode].N)
	envVars, err := dockerMachineEnv(settings, evt)
	if err != nil {
		return
	}
	envVars = dockerMachineNodeEnv(envVars, evt.ID(), machineHomeDir, evt.State[firstNode])

	command := []string{dmBin, "ssh", evt.State[firstNode].ID(), "docker", "run", "-d", "--volume=/home/docker/nodeconfig:/payload/config", "-p", "1317:1317", "apeunit/launchpayload", "/payload/runlightclient.sh", evt.State[firstNode].Instance.IPAddress, evt.ID()}
	// command = []string{"scp", evt.Payload.CLIPath, fmt.Sprintf("%s:/home/docker", evt.State[firstNode].ID())}
	log.Debugf("Running docker-machine %s on validator %s machine; envVars %s\n", command, firstNode, envVars)
	_, err = cmdRunner(command, envVars)
	if err != nil {
		log.Error(err)
		return
	}

	log.Infoln("Copying the faucet account and configuration to the first validator machine")
	faucetAccount := evt.FaucetAccount()
	v, _ := evt.Validators()
	command = []string{dmBin, "scp", "-r", faucetAccount.ConfigLocation.CLIConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig/faucet_account", evt.State[v[0]].ID())}
	_, err = cmdRunner(command, envVars)
	if err != nil {
		log.Errorf("docker-machine %s failed with %s", command, err)
		return
	}
	// docker-machine chmod -R 777 /home/docker/nodeconfig AGAIN - what a mess!
	command = []string{dmBin, "ssh", evt.State[v[0]].ID(), "chmod", "-R", "777", "/home/docker/nodeconfig"}
	_, err = cmdRunner(command, envVars)
	if err != nil {
		log.Errorf("docker-machine %s failed with %s", command, err)
		return
	}

	evtDir, err := evts(settings, evt.ID())
	if err != nil {
		log.Error(err)
		return
	}

	command = []string{dmBin, "scp", "-r", filepath.Join(evtDir, "faucetconfig.yml"), fmt.Sprintf("%s:/home/docker/nodeconfig/", evt.State[v[0]].ID())}
	_, err = cmdRunner(command, envVars)
	if err != nil {
		log.Errorf("docker-machine %s failed with %s", command, err)
		return
	}

	log.Infoln("Starting the faucet")
	firstValidator := evt.State[v[0]]
	envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(firstValidator.N), firstValidator)
	command = []string{"docker", "run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "8000:8000", evt.Payload.DockerImage, "/payload/gofaucet", "/payload/config/faucetconfig.yml"}
	log.Debugf("Running docker %s on %s; envVars %s\n", command, firstValidator.ID(), envVars)
	_, err = cmdRunner(command, envVars)
	if err != nil {
		log.Errorf("docker %s failed with %s", args, err)
		return
	}
	return
}
