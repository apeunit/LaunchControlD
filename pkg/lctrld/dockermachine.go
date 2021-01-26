package lctrld

import (
	"fmt"
	"os"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
)

// dockerMachineEnv ensures we are talking to the correct docker-machine binary, and that the context is the eventivize workspace directory
func dockerMachineEnv(settings config.Schema, evt *model.Event) (env []string, err error) {
	// add extra PATHs to find other docker-machine binaries
	p := append([]string{}, bin(settings, ""), os.Getenv("PATH"))
	envPath := fmt.Sprintf("PATH=%s", strings.Join(p, ":"))

	// set MACHINE_STORAGE_PATH
	home, err := evts(settings, evt.ID()) // this gives you the relative path to the event home
	if err != nil {
		return
	}
	envMachineStoragePath := fmt.Sprintf("MACHINE_STORAGE_PATH=%s", _path(home, ".docker", "machine"))

	// add docker-machine driver env vars
	env = settings.DockerMachine.Drivers[evt.Provider].Env

	env = append(env, envMachineStoragePath, envPath)
	env = append(env, settings.DockerMachine.Env...)
	return
}

// dockerMachineNodeEnv recreates the output of docker-machine env <MACHINE NAME>, to run a command inside the docker-machine provisioned node.
func dockerMachineNodeEnv(envVars []string, eventID, machineHomeDir string, state *model.MachineConfig) []string {
	envVars = append(
		envVars,
		"DOCKER_TLS_VERIFY=1",
		fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", state.Instance.IPAddress),
		fmt.Sprintf("DOCKER_CERT_PATH=%s", machineHomeDir),
		fmt.Sprintf("DOCKER_MACHINE_NAME=%s", state.ID()),
	)
	return envVars
}

// DockerMachineInterface is a mocking interface for functions that need to
// read docker-machine config files
type DockerMachineInterface interface {
	HomeDir(string) string
	ReadConfig(string) (*model.MachineConfig, error)
}

// DockerMachineConfig holds information that lets lctrld read the state of a
// docker-machine provisioning
type DockerMachineConfig struct {
	EventID  string
	Settings config.Schema
}

// NewDockerMachineConfig ensures that all fields of a DockerMachineConfig are filled out
func NewDockerMachineConfig(settings config.Schema, eventID string) *DockerMachineConfig {
	return &DockerMachineConfig{
		EventID:  eventID,
		Settings: settings,
	}
}

// HomeDir returns the path of a docker-machine instance home, e.g.
// /tmp/workspace/evts/drop-xxx/.docker/machine/machines/drop-xxx-0/
func (dmc *DockerMachineConfig) HomeDir(machineN string) string {
	return _path(dmc.Settings.Workspace, evtsDir, dmc.EventID, ".docker", "machine", "machines", fmt.Sprintf("%s-%s", dmc.EventID, machineN))
}

// ReadConfig return configuration of a docker machine
func (dmc *DockerMachineConfig) ReadConfig(machineN string) (mc *model.MachineConfig, err error) {
	mc = new(model.MachineConfig)
	dmcf := new(DockerMachineConfigFormat)
	err = utils.LoadJSON(_path(dmc.HomeDir(machineN), "config.json"), &dmcf)
	if err != nil {
		return nil, err
	}
	mc.Instance.IPAddress = dmcf.Driver.IPAddress
	mc.Instance.MachineName = dmcf.Driver.MachineName
	mc.Instance.SSHUser = dmcf.Driver.SSHUser
	mc.Instance.SSHPort = dmcf.Driver.SSHPort
	mc.Instance.SSHKeyPath = dmcf.Driver.SSHKeyPath
	mc.Instance.StorePath = dmcf.Driver.StorePath
	mc.N = strings.Split(dmcf.Name, "-")[2]
	mc.EventID = dmc.EventID
	return
}

// DockerMachineConfigFormat is the structure of
// .docker/machine/machines/<MACHINE NAME>/config.json, which describes a deployed VM's configuration
type DockerMachineConfigFormat struct {
	ConfigVersion int `json:"ConfigVersion"`
	Driver        struct {
		IPAddress      string `json:"IPAddress"`
		MachineName    string `json:"MachineName"`
		SSHUser        string `json:"SSHUser"`
		SSHPort        int    `json:"SSHPort"`
		SSHKeyPath     string `json:"SSHKeyPath"`
		StorePath      string `json:"StorePath"`
		SwarmMaster    bool   `json:"SwarmMaster"`
		SwarmHost      string `json:"SwarmHost"`
		SwarmDiscovery string `json:"SwarmDiscovery"`
		VBoxManager    struct {
		} `json:"VBoxManager"`
		HostInterfaces struct {
		} `json:"HostInterfaces"`
		CPU                 int    `json:"CPU"`
		Memory              int    `json:"Memory"`
		DiskSize            int    `json:"DiskSize"`
		NatNicType          string `json:"NatNicType"`
		Boot2DockerURL      string `json:"Boot2DockerURL"`
		Boot2DockerImportVM string `json:"Boot2DockerImportVM"`
		HostDNSResolver     bool   `json:"HostDNSResolver"`
		HostOnlyCIDR        string `json:"HostOnlyCIDR"`
		HostOnlyNicType     string `json:"HostOnlyNicType"`
		HostOnlyPromiscMode string `json:"HostOnlyPromiscMode"`
		UIType              string `json:"UIType"`
		HostOnlyNoDHCP      bool   `json:"HostOnlyNoDHCP"`
		NoShare             bool   `json:"NoShare"`
		DNSProxy            bool   `json:"DNSProxy"`
		NoVTXCheck          bool   `json:"NoVTXCheck"`
		ShareFolder         string `json:"ShareFolder"`
	} `json:"Driver"`
	DriverName  string `json:"DriverName"`
	HostOptions struct {
		Driver        string `json:"Driver"`
		Memory        int    `json:"Memory"`
		Disk          int    `json:"Disk"`
		EngineOptions struct {
			ArbitraryFlags   []interface{} `json:"ArbitraryFlags"`
			DNS              interface{}   `json:"Dns"`
			GraphDir         string        `json:"GraphDir"`
			Env              []interface{} `json:"Env"`
			Ipv6             bool          `json:"Ipv6"`
			InsecureRegistry []interface{} `json:"InsecureRegistry"`
			Labels           []interface{} `json:"Labels"`
			LogLevel         string        `json:"LogLevel"`
			StorageDriver    string        `json:"StorageDriver"`
			SelinuxEnabled   bool          `json:"SelinuxEnabled"`
			TLSVerify        bool          `json:"TlsVerify"`
			RegistryMirror   []interface{} `json:"RegistryMirror"`
			InstallURL       string        `json:"InstallURL"`
		} `json:"EngineOptions"`
		SwarmOptions struct {
			IsSwarm            bool          `json:"IsSwarm"`
			Address            string        `json:"Address"`
			Discovery          string        `json:"Discovery"`
			Agent              bool          `json:"Agent"`
			Master             bool          `json:"Master"`
			Host               string        `json:"Host"`
			Image              string        `json:"Image"`
			Strategy           string        `json:"Strategy"`
			Heartbeat          int           `json:"Heartbeat"`
			Overcommit         int           `json:"Overcommit"`
			ArbitraryFlags     []interface{} `json:"ArbitraryFlags"`
			ArbitraryJoinFlags []interface{} `json:"ArbitraryJoinFlags"`
			Env                interface{}   `json:"Env"`
			IsExperimental     bool          `json:"IsExperimental"`
		} `json:"SwarmOptions"`
		AuthOptions struct {
			CertDir              string        `json:"CertDir"`
			CaCertPath           string        `json:"CaCertPath"`
			CaPrivateKeyPath     string        `json:"CaPrivateKeyPath"`
			CaCertRemotePath     string        `json:"CaCertRemotePath"`
			ServerCertPath       string        `json:"ServerCertPath"`
			ServerKeyPath        string        `json:"ServerKeyPath"`
			ClientKeyPath        string        `json:"ClientKeyPath"`
			ServerCertRemotePath string        `json:"ServerCertRemotePath"`
			ServerKeyRemotePath  string        `json:"ServerKeyRemotePath"`
			ClientCertPath       string        `json:"ClientCertPath"`
			ServerCertSANs       []interface{} `json:"ServerCertSANs"`
			StorePath            string        `json:"StorePath"`
		} `json:"AuthOptions"`
	} `json:"HostOptions"`
	Name string `json:"Name"`
}
