package lctrld

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
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
	_, validatorAccounts := evt.Validators()
	for i := range validatorAccounts {
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
	evt, err := model.LoadEvtvzE(p)
	if err != nil {
		return
	}
	// run the rm command for each validator
	dmBin := dmBin(settings)
	// set the path to find the executable
	envVars, err := dockerEnv(settings, *evt)
	if err != nil {
		return
	}

	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		host := evt.NodeID(i)
		//driver := settings.DockerMachine.Drivers[evt.Provider]
		log.Infof("Node ID for %s is %s", v, host)
		// create the parameters
		out, err := runCommand(dmBin, []string{"stop", host}, envVars)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(host, "stop:", out)

		out, err = runCommand(dmBin, []string{"rm", host}, envVars)
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
func Provision(settings config.Schema, evtID string) (err error) {
	evt, err := loadEvent(settings, evtID)
	if err != nil {
		return
	}
	// Outputter
	var out []byte
	dmBin := dmBin(settings)
	// set the path to find the executable
	evnVars, err := dockerEnv(settings, evt)
	if err != nil {
		return
	}
	// init docker nodes map
	evt.State = make(map[string]*model.MachineConfig)
	// run the thing
	_, validatorAccounts := evt.Validators()
	for i, v := range validatorAccounts {
		host := evt.NodeID(i)
		driver := settings.DockerMachine.Drivers[evt.Provider]

		log.Infof("Node ID for %s is %s", v, host)
		// create the parameters
		p := []string{"create", "--driver", evt.Provider}
		p = append(p, driver.Params...)
		p = append(p, host)
		log.Debug("Provision cmd: ", dmBin, evt.Provider, host)
		/// prepare the command
		cmd := exec.Command(dmBin, p...)
		// add the binary folder to the exec path
		cmd.Env = evnVars
		log.Debug("Provision env vars set to ", cmd.Env)
		// execute the command
		out, err = cmd.CombinedOutput()
		if err != nil {
			log.Errorf("Provision cmd failed with %s, %s\n", err, out)
			break
		}
		log.Debug("Provision cmd output: ", string(out), err)
		// load the configuration of the machine
		mc, err := machineConfig(settings, evt.ID(), i)
		if err != nil {
			log.Errorf("Provision read machine config error:", err)
			break
		}

		ip, err := runCommand(dmBin, []string{"ip", host}, evnVars)
		if err != nil {
			log.Error(err)
			break
		}
		mc.Instance.IPAddress = string(ip)

		evt.State[v.Name] = &mc
	}
	if err != nil {
		return
	}
	err = storeEvent(settings, evt)
	return
}

func DeployPayload(settings config.Schema, evtID string) (err error) {
	evt, err := loadEvent(settings, evtID)
	if err != nil {
		return
	}
	dmBin := dmBin(settings)

	fmt.Println("Copying node configs to each provisioned machine")
	for _, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerEnv() failed while generating envVars: %s", err)
			break
		}

		pathDaemon, pathCLI, err := getNodeConfigDir(settings, evtID, state.ID)
		if err != nil {
			log.Errorf("Error while getting Cosmos-SDK cli/daemon config directory: %s", err)
			break
		}

		// docker-machine ssh mkdir -p /home/docker/nodeconfig
		args := []string{"ssh", state.ID, "mkdir", "-p", "/home/docker/nodeconfig"}
		_, err = runCommand(dmBin, args, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", args, err)
			break
		}

		// docker-machine scp -r pathDaemon evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		args = []string{"scp", "-r", pathDaemon, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID)}
		_, err = runCommand(dmBin, args, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", args, err)
			break
		}

		// docker-machine scp -r pathCLI evtx-d97517a3673688070aef-0:/home/docker/nodeconfig
		args = []string{"scp", "-r", pathCLI, fmt.Sprintf("%s:/home/docker/nodeconfig", state.ID)}
		_, err = runCommand(dmBin, args, envVars)
		if err != nil {
			log.Errorf("docker-machine %s failed with %s", args, err)
			break
		}
	}

	for email, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		machineID, err := state.NumberID()
		if err != nil {
			log.Errorf("state.NumberID() failed with %s", err)
			break
		}

		machineHomeDir := machineHome(settings, evtID, machineID)
		envVars = append(envVars, "DOCKER_TLS_VERIFY=1", fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", state.Instance.IPAddress), fmt.Sprintf("DOCKER_CERT_PATH=%s", machineHomeDir), fmt.Sprintf("DOCKER_MACHINE_NAME=%s-%s", evtID, state.ID))

		// in docker-machine provisioned machine: docker pull apeunit/launchpayload
		args := []string{"pull", evt.DockerImage}
		fmt.Printf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		out, err := runCommand("docker", args, envVars)
		if err != nil {
			log.Errorf("docker %s failed with %s", args, err)
			break
		}
		fmt.Println("docker:", out)
	}

	for email, state := range evt.State {
		envVars, err := dockerEnv(settings, evt)
		if err != nil {
			log.Errorf("dockerEnv() failed while generating envVars: %s", err)
			break
		}

		// Build the output of docker-machine -s /tmp/workspace/evts/evtx-d97517a3673688070aef/.docker/machine/ env evtx-d97517a3673688070aef-1
		machineID, err := state.NumberID()
		if err != nil {
			log.Errorf("state.NumberID() failed with %s", err)
			break
		}

		machineHomeDir := machineHome(settings, evtID, machineID)
		envVars = append(envVars, "DOCKER_TLS_VERIFY=1", fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", state.Instance.IPAddress), fmt.Sprintf("DOCKER_CERT_PATH=%s", machineHomeDir), fmt.Sprintf("DOCKER_MACHINE_NAME=%s-%s", evtID, state.ID))

		// in docker-machine provisioned machine: docker run -v /home/docker/nodeconfig:/payload/config apeunit/launchpayload
		args := []string{"run", "-d", "-v", "/home/docker/nodeconfig:/payload/config", "-p", "26656:26656", "-p", "26657:26657", "-p", "26658:26658", evt.DockerImage}
		fmt.Printf("Running docker %s for validator %s machine; envVars %s\n", args, email, envVars)
		out, err := runCommand("docker", args, envVars)
		if err != nil {
			log.Errorf("docker %s failed with %s", args, err)
			break
		}
		fmt.Println("docker:", out)
	}
	return
}
