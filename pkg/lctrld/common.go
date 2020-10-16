package lctrld

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	log "github.com/sirupsen/logrus"
)

const (
	binDir            = "bin"
	tmpDir            = "tmp"
	evtsDir           = "evts"
	evtDescriptorFile = "event.json"
	dockerHome        = ".docker"
)

func bin(settings config.Schema, file string) string {
	return _path(settings.Workspace, binDir, file)
}

func tmp(settings config.Schema) (string, error) {
	return ioutil.TempDir(_path(settings.Workspace, tmpDir), "")
}

func evts(settings config.Schema, dir string) (string, error) {
	return _absPath(_path(settings.Workspace, evtsDir, dir))
}

// DockerMachineInterface is sa mocking interface for functions that need to
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

// HomeDir get the path of a docker-machine instance home, e.g.
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

// DockerMachineConfigFormat is the structure of a JSON file outputted by
// docker-machine
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

// evtDescriptor returns the absolute path to the event descriptor file
func evtDescriptor(settings config.Schema, evtID string) (path string, err error) {
	path, err = evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	return
}

//LoadEvent returns the Event model of the specified event ID
func LoadEvent(settings config.Schema, evtID string) (evt *model.EvtvzE, err error) {
	path, err := evts(settings, evtID)
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.LoadJSON(path, &evt)
	return
}

// StoreEvent returns the Event model of the specified event ID
func StoreEvent(settings config.Schema, evt *model.EvtvzE) (err error) {
	path, err := evts(settings, evt.ID())
	if err != nil {
		return
	}
	path = _path(path, evtDescriptorFile)
	err = utils.StoreJSON(path, evt)
	return
}

func _path(pieces ...string) string {
	return filepath.Join(pieces...)
}

func _absPath(relative string) (string, error) {
	return filepath.Abs(relative)
}

func dmBin(settings config.Schema) string {
	return bin(settings, settings.DockerMachine.Binary)
}

func dmDriverBin(settings config.Schema, driverName string) string {
	return bin(settings, settings.DockerMachine.Drivers[driverName].Binary)
}

// CommandRunner func type allows for mocking out RunCommand()
type CommandRunner func(string, []string, []string) (string, error)

// RunCommand runs a command
func RunCommand(bin string, args, envVars []string) (out string, err error) {
	/// prepare the command
	cmd := exec.Command(bin, args...)
	// add the binary folder to the exec path
	cmd.Env = envVars
	log.Debug("command env vars set to ", cmd.Env)
	// execute the command
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("%s %s failed with %s, %s\n", bin, args, err, string(o))
		return
	}
	out = strings.TrimSpace(string(o))
	log.Debug("command stdout: ", out)
	return
}
