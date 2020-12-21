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
		log.Fatal("Inspect failed:", err)
		return
	}
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerMachineEnv(settings, evt)
	_, validatorAccounts := evt.Validators()
	for i := range validatorAccounts {
		host := evt.NodeID(i)
		out, err := cmdRunner(dmBin, []string{"status", host}, envVars)
		if err != nil {
			break
		}
		fmt.Println(host, "status:", out)
		out, err = cmdRunner(dmBin, []string{"ip", host}, envVars)
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
		out, err := cmdRunner(dmBin, []string{"stop", host}, envVars)
		if err != nil {
			// TODO: preferr logging over printing to stdout
			fmt.Println(err)
		}
		// TODO: prefer logging over printing on stdout
		fmt.Println(host, "stop:", out)

		out, err = cmdRunner(dmBin, []string{"rm", host}, envVars)
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
func Provision(settings config.Schema, evt *model.Event, cmdRunner CommandRunner, dmc DockerMachineInterface) error {
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
		p := []string{"--debug", "create", "--driver", evt.Provider}
		p = append(p, driver.Params...)
		p = append(p, host)

		log.Debugf("Provision cmd: %s %s %s", dmBin, evt.Provider, host)
		log.Debug("Provision env vars set to ", envVars)
		out, err := cmdRunner(dmBin, p, envVars)
		if err != nil {
			log.Fatalf("Provision cmd failed with %s, %s\n", err, out)
			break
		}

		log.Debug("Provision cmd output: ", string(out), err)
		// load the configuration of the machine
		mc, err := dmc.ReadConfig(fmt.Sprint(i))
		if err != nil {
			log.Fatal("Provision read machine config error:", err)
			break
		}
		evt.State[v.Name] = mc
	}
	if err != nil {
		return err
	}
	log.Infof("Your event ID is %s", evt.ID())
	return nil
}

// RereadDockerMachineInfo is useful when docker-machine failed during 'create',
// and a human fixed the problem, and wants to continue
func RereadDockerMachineInfo(settings config.Schema, evt *model.Event, dmc DockerMachineInterface) (event *model.Event, err error) {
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		mc, err := dmc.ReadConfig(fmt.Sprint(i))
		if err != nil {
			log.Fatal("Provision read machine config error:", err)
			break
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
			log.Fatalf("dockerMachineEnv() failed while generating envVars: %s", err)
			break
		}

		// docker-machine ssh mkdir -p /home/docker/nodeconfig
		args := []string{"ssh", state.ID(), "mkdir", "-p", "/home/docker/nodeconfig"}
		_, err = cmdRunner(dmBin, args, envVars)
		if err != nil {
			log.Fatalf("docker-machine %s failed with %s", args, err)
			break
		}

		// docker-machine scp -r pathDaemon evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		args = []string{"scp", "-r", evt.Accounts[name].ConfigLocation.DaemonConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID())}
		_, err = cmdRunner(dmBin, args, envVars)
		if err != nil {
			log.Fatalf("docker-machine %s failed with %s", args, err)
			break
		}

		// docker-machine scp -r pathCLI evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		args = []string{"scp", "-r", evt.Accounts[name].ConfigLocation.CLIConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID())}
		_, err = cmdRunner(dmBin, args, envVars)
		if err != nil {
			log.Fatalf("docker-machine %s failed with %s", args, err)
			break
		}

		// docker-machine chmod -R 777 /home/docker/nodeconfig
		args = []string{"ssh", state.ID(), "chmod", "-R", "777", "/home/docker/nodeconfig"}
		_, err = cmdRunner(dmBin, args, envVars)
		if err != nil {
			log.Fatalf("docker-machine %s failed with %s", args, err)
			break
		}
	}

	log.Infof("Running docker pull %s on each provisioned machine", evt.Payload.DockerImage)
	for email, state := range evt.State {
		envVars, err := dockerMachineEnv(settings, evt)
		if err != nil {
			log.Fatalf("dockerMachineEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(state.N), state)

		// in docker-machine provisioned machine: docker pull apeunit/launchpayload
		args := []string{"pull", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		_, err = cmdRunner("docker", args, envVars)
		if err != nil {
			log.Fatalf("docker %s failed with %s", args, err)
			break
		}
	}

	log.Infoln("Running the dockerized Cosmos daemons on the provisioned machines")
	for email, state := range evt.State {
		envVars, err := dockerMachineEnv(settings, evt)
		if err != nil {
			log.Fatalf("dockerMachineEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(state.N), state)

		// in docker-machine provisioned machine: docker run -v /home/docker/nodeconfig:/payload/config apeunit/launchpayload
		args := []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "26656:26656", "-p", "26657:26657", "-p", "26658:26658", evt.Payload.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		_, err = cmdRunner("docker", args, envVars)
		if err != nil {
			log.Fatalf("docker %s failed with %s", args, err)
			break
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

	args = []string{"ssh", evt.State[firstNode].ID(), "docker", "run", "-d", "--volume=/home/docker/nodeconfig:/payload/config", "-p", "1317:1317", "apeunit/launchpayload", "/payload/runlightclient.sh", evt.State[firstNode].Instance.IPAddress, evt.ID()}
	// args = []string{"scp", evt.Payload.CLIPath, fmt.Sprintf("%s:/home/docker", evt.State[firstNode].ID())}
	log.Debugf("Running docker-machine %s on validator %s machine; envVars %s\n", args, firstNode, envVars)
	_, err = cmdRunner(dmBin, args, envVars)
	if err != nil {
		log.Fatal(err)
		return
	}

	// args = []string{"ssh", evt.State[firstNode].ID(), "/home/docker/launchpayloadcli", "rest-server", "--laddr", "tcp://0.0.0.0:1317", "--node", fmt.Sprintf("tcp://%s:26657", evt.State[firstNode].Instance.IPAddress), "--unsafe-cors", "--chain-id", evt.ID(), "--home", "/home/docker/nodeconfig/cli", "&"}
	// o, err = cmdRunner(dmBin, args, envVars)
	// log.Infoln(o)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return
	// }

	log.Infoln("Copying the faucet account and configuration to the first validator machine")
	faucetAccount := evt.FaucetAccount()
	v, _ := evt.Validators()
	args = []string{"scp", "-r", faucetAccount.ConfigLocation.CLIConfigDir, fmt.Sprintf("%s:/home/docker/nodeconfig/faucet_account", evt.State[v[0]].ID())}
	_, err = cmdRunner("docker-machine", args, envVars)
	if err != nil {
		log.Fatalf("docker-machine %s failed with %s", args, err)
		return
	}
	// docker-machine chmod -R 777 /home/docker/nodeconfig AGAIN - what a mess!
	args = []string{"ssh", evt.State[v[0]].ID(), "chmod", "-R", "777", "/home/docker/nodeconfig"}
	_, err = cmdRunner(dmBin, args, envVars)
	if err != nil {
		log.Fatalf("docker-machine %s failed with %s", args, err)
	}

	evtDir, err := evts(settings, evt.ID())
	if err != nil {
		log.Fatal(err)
	}

	args = []string{"scp", "-r", filepath.Join(evtDir, "faucetconfig.yml"), fmt.Sprintf("%s:/home/docker/nodeconfig/", evt.State[v[0]].ID())}
	_, err = cmdRunner("docker-machine", args, envVars)
	if err != nil {
		log.Fatalf("docker-machine %s failed with %s", args, err)
		return
	}

	log.Infoln("Starting the faucet")
	firstValidator := evt.State[v[0]]
	envVars = dockerMachineNodeEnv(envVars, evt.ID(), dmc.HomeDir(firstValidator.N), firstValidator)
	args = []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "8000:8000", evt.Payload.DockerImage, "/payload/gofaucet", "/payload/config/faucetconfig.yml"}
	log.Debugf("Running docker %s on %s; envVars %s\n", args, firstValidator.ID(), envVars)
	_, err = cmdRunner("docker", args, envVars)
	if err != nil {
		log.Fatalf("docker %s failed with %s", args, err)
		return
	}
	return
}
