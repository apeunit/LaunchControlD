package lctrld

import (
	"fmt"
	"os"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"
)

func dockerEnv(settings config.Schema, evt *model.EvtvzE) (env []string, err error) {
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
func InspectEvent(settings config.Schema, evt *model.EvtvzE, cmdRunner CommandRunner) (err error) {
	path, err := evts(settings, evt.ID())
	log.Debugln("InspectEvent event", evt.ID(), "home:", path)
	if err != nil {
		log.Fatal("Inspect failed:", err)
		return
	}
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerEnv(settings, evt)
	_, validatorAccounts := evt.Validators()
	for i := range validatorAccounts {
		host := evt.NodeID(i)
		out, err := cmdRunner(dmBin, []string{"status", host}, envVars)
		if err != nil {
			break
		}
		fmt.Println(host, "status:", out)
		out, err = cmdRunner(dmBin, []string{"ip", host}, envVars)
		fmt.Println(host, "IP:", out)
	}
	return
}

// DestroyEvent destroy an existing event
func DestroyEvent(settings config.Schema, evt *model.EvtvzE, cmdRunner CommandRunner) (err error) {
	path, err := evts(settings, evt.ID())
	log.Debugln("DestroyEvent event", evt.ID(), "home:", path)
	if err != nil {
		log.Fatal("DestroyEvent failed:", err)
		return
	}
	if !utils.FileExists(path) {
		err = fmt.Errorf("Event ID %s not found", evt.ID())
		log.Fatal("DestroyEvent failed:", err)
		return
	}
	// load the descriptor
	p, err := evtDescriptor(settings, evt.ID())
	log.Debug("DestroyEvent event descriptor:", p)
	if err != nil {
		log.Fatal("DestroyEvent failed:", err)
		return
	}

	// run the rm command for each validator
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerEnv(settings, evt)
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
			fmt.Println(err)
		}
		fmt.Println(host, "stop:", out)

		out, err = cmdRunner(dmBin, []string{"rm", host}, envVars)
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

// Provision provision the infrastructure for the event
func Provision(settings config.Schema, evt *model.EvtvzE, cmdRunner CommandRunner, dmc DockerMachineInterface) (*model.EvtvzE, error) {
	// Outputter
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerEnv(settings, evt)
	if err != nil {
		return nil, err
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
		p := []string{"create", "--driver", evt.Provider}
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

		ip, err := cmdRunner(dmBin, []string{"ip", host}, envVars)
		if err != nil {
			log.Fatal(err)
			break
		}
		mc.Instance.IPAddress = string(ip)
		mc.N = fmt.Sprint(i)
		mc.EventID = evt.ID()
		evt.State[v.Name] = mc
	}
	if err != nil {
		return nil, err
	}
	log.Infof("Your event ID is %s", evt.ID());
	return evt, nil
}

// DeployPayload tells the provisioned machines to run the configured docker
// image
func DeployPayload(settings config.Schema, evt *model.EvtvzE, cmdRunner CommandRunner, dmc DockerMachineInterface) (err error) {
	dmBin := dmBin(settings)

	log.Infoln("Copying node configs to each provisioned machine")
	for name, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Fatalf("dockerEnv() failed while generating envVars: %s", err)
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

	log.Infof("Running docker pull %s on each provisioned machine", evt.DockerImage)
	for email, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Fatalf("dockerEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		machineHomeDir := dmc.HomeDir(state.N)
		envVars = append(envVars, "DOCKER_TLS_VERIFY=1", fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", state.Instance.IPAddress), fmt.Sprintf("DOCKER_CERT_PATH=%s", machineHomeDir), fmt.Sprintf("DOCKER_MACHINE_NAME=%s-%s", evt.ID(), state.ID()))

		// in docker-machine provisioned machine: docker pull apeunit/launchpayload
		args := []string{"pull", evt.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		_, err = cmdRunner("docker", args, envVars)
		if err != nil {
			log.Fatalf("docker %s failed with %s", args, err)
			break
		}
	}

	log.Infoln("Running the docker image on the provisioned machines")
	for email, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Fatalf("dockerEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		machineHomeDir := dmc.HomeDir(state.N)
		envVars = append(envVars, "DOCKER_TLS_VERIFY=1", fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", state.Instance.IPAddress), fmt.Sprintf("DOCKER_CERT_PATH=%s", machineHomeDir), fmt.Sprintf("DOCKER_MACHINE_NAME=%s-%s", evt.ID(), state.ID()))

		// in docker-machine provisioned machine: docker run -v /home/docker/nodeconfig:/payload/config apeunit/launchpayload
		args := []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "26656:26656", "-p", "26657:26657", "-p", "26658:26658", evt.DockerImage}
		log.Debugf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		_, err = cmdRunner("docker", args, envVars)
		if err != nil {
			log.Fatalf("docker %s failed with %s", args, err)
			break
		}
	}
	return
}
