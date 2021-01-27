package lctrld

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
)

// DockerMachine implements docker-machine functionality for lctrld
type DockerMachine struct {
	EventID  string
	EnvVars  []string
	Settings config.Schema
}

// NewDockerMachine ensures that all fields of a DockerMachineConfig are filled out
func NewDockerMachine(settings config.Schema, eventID string) *DockerMachine {
	envVars := utils.BuildEnvVars(settings)

	// set MACHINE_STORAGE_PATH
	home, err := utils.Evts(settings, eventID) // this gives you the relative path to the event home
	if err != nil {
		return nil
	}
	envMachineStoragePath := fmt.Sprintf("MACHINE_STORAGE_PATH=%s", filepath.Join(home, ".docker", "machine"))
	envVars = append(envVars, envMachineStoragePath)

	// include extra environment variables from the lctrld settings
	envVars = append(envVars, settings.DockerMachine.Env...)
	return &DockerMachine{
		EventID:  eventID,
		EnvVars:  envVars,
		Settings: settings,
	}
}

// HomeDir returns the path of a docker-machine instance home, e.g.
// /tmp/workspace/evts/drop-xxx/.docker/machine/machines/drop-xxx-0/
func (dm *DockerMachine) HomeDir(machineName string) string {
	return filepath.Join(dm.Settings.Workspace, utils.EvtsDir, dm.EventID, ".docker", "machine", "machines", machineName)
}

// ReadConfig return configuration of a docker machine
func (dm *DockerMachine) ReadConfig(machineName string) (mc *model.Machine, err error) {
	mc = new(model.Machine)
	dmcf := new(DockerMachineConfigFormat)
	err = utils.LoadJSON(filepath.Join(dm.HomeDir(machineName), "config.json"), &dmcf)
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
	mc.EventID = dm.EventID
	return
}

// ProvisionMachine runs docker-machine create MACHINE_NAME
func (dm *DockerMachine) ProvisionMachine(machineName, provider string, cmdRunner cmdrunner.CommandRunner) (mc *model.Machine, err error) {
	driver := dm.Settings.DockerMachine.Drivers[provider]

	p := []string{utils.DmBin(dm.Settings), "--debug", "create", "--driver", provider, "--engine-install-url", "https://releases.rancher.com/install-docker/19.03.9.sh"}
	p = append(p, driver.Params...)
	p = append(p, machineName)

	_, err = cmdRunner(p, dm.EnvVars)
	if err != nil {
		return
	}

	return dm.ReadConfig(machineName)
}

// StopMachine runs docker-machine stop MACHINE_NAME && docker-machine rm -y MACHINE_NAME
func (dm *DockerMachine) StopMachine(machineName string, cmdRunner cmdrunner.CommandRunner) (err error) {
	p := []string{utils.DmBin(dm.Settings), "stop", machineName}
	_, err = cmdRunner(p, dm.EnvVars)
	if err != nil {
		return
	}

	p = []string{utils.DmBin(dm.Settings), "rm", "-y", machineName}
	_, err = cmdRunner(p, dm.EnvVars)
	if err != nil {
		return
	}
	return
}

// Status runs docker-machine ip MACHINE_NAME and docker-machine status machine_NAME
func (dm *DockerMachine) Status(machineName string, cmdRunner cmdrunner.CommandRunner) (out string, err error) {
	var out1, out2 string
	p := []string{utils.DmBin(dm.Settings), "status", machineName}
	out1, err = cmdRunner(p, dm.EnvVars)
	if err != nil {
		return
	}
	p = []string{utils.DmBin(dm.Settings), "ip", machineName}
	out2, err = cmdRunner(p, dm.EnvVars)
	if err != nil {
		return
	}
	out = strings.Join([]string{out1, out2}, "\n")
	return
}

// RunDocker tells this computer's docker binary to talk with the remote
// machine's docker installation and run a commnad for safety, this command
// prepends "docker" to any command you send it. Therefore, to run "docker pull
// <IMAGE>" on the remote machine, pass in []string{"pull", IMAGENAME}
func (dm *DockerMachine) RunDocker(machineName string, cmd []string, cmdRunner cmdrunner.CommandRunner) (out string, err error) {
	ip, err := cmdRunner([]string{utils.DmBin(dm.Settings), "ip", machineName}, dm.EnvVars)
	if err != nil {
		return
	}

	envVars := append(dm.EnvVars,
		"DOCKER_TLS_VERIFY=1",
		fmt.Sprintf("DOCKER_HOST=tcp://%s:2376", ip),
		fmt.Sprintf("DOCKER_CERT_PATH=%s", dm.HomeDir(machineName)),
		fmt.Sprintf("DOCKER_MACHINE_NAME=%s", machineName),
	)
	finalCmd := []string{"docker"}
	finalCmd = append(finalCmd, cmd...)
	out, err = cmdRunner(finalCmd, envVars)
	return
}

// Run uses docker-machine ssh to run a command on the remote machine.
func (dm *DockerMachine) Run(machineName string, cmd []string, cmdRunner cmdrunner.CommandRunner) (out string, err error) {
	finalCmd := []string{utils.DmBin(dm.Settings), "ssh", machineName}
	finalCmd = append(finalCmd, cmd...)
	out, err = cmdRunner(finalCmd, dm.EnvVars)
	return
}

// Copy recursively copies a path from the local machine to the provisioned Machine
func (dm *DockerMachine) Copy(machineName, sourcePath, destPath string, cmdRunner cmdrunner.CommandRunner) (err error) {
	p := []string{utils.DmBin(dm.Settings), "scp", "-r", sourcePath, fmt.Sprintf("%s:%s", machineName, destPath)}
	_, err = cmdRunner(p, dm.EnvVars)
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
