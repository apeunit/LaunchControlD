package lctrld

import (
	"fmt"
	"os"
	"testing"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	"github.com/stretchr/testify/assert"
)

var mockSettings = config.Schema{
	Workspace: "/tmp/workspace",
	DockerMachine: config.DockerMachine{
		Version:   "0.16.2",
		BinaryURL: "https://github.com/docker/machine/releases/download/v0.16.2/docker-machine-Linux-x86_64",
		Binary:    "docker-machine",
		Drivers: map[string]config.DockerMachineDriver{
			"": {
				Version:   "",
				BinaryURL: "",
				Binary:    "",
				Params:    nil,
				Env:       nil,
			},
		},
		Env: []string{"VIRTUALBOX_BOOT2DOCKER_URL=/home/shinichi/boot2docker.iso"},
	},
}

func TestDockerMachineConfig(t *testing.T) {
	dmc := NewDockerMachine(mockSettings, "drop-28b10d4eff415a7b0b2c")
	assert.Equal(t, "testdata/evts/drop-28b10d4eff415a7b0b2c/.docker/machine/machines/drop-28b10d4eff415a7b0b2c-0", dmc.HomeDir("0"))

	mc, err := dmc.ReadConfig("0")
	assert.Nil(t, err)
	mcExpected := &model.Machine{
		N:                "0",
		EventID:          "drop-28b10d4eff415a7b0b2c",
		DriverName:       "",
		TendermintNodeID: "",
		Instance: model.MachineNetworkConfig{
			IPAddress:   "192.168.99.100",
			MachineName: "drop-28b10d4eff415a7b0b2c-0",
			SSHUser:     "docker",
			SSHPort:     36027,
			SSHKeyPath:  "/tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/.docker/machine/machines/drop-28b10d4eff415a7b0b2c-0/id_rsa",
			StorePath:   "/tmp/workspace/evts/drop-28b10d4eff415a7b0b2c/.docker/machine",
		},
	}
	assert.Equal(t, mcExpected, mc)
}

func TestDockerMachineProvisionMachine(t *testing.T) {
	evtID := "test-2849deadbeef0293"
	machineName := fmt.Sprintf("%s-%s", evtID, "0")
	dm := NewDockerMachine(mockSettings, evtID)
	fmt.Println(dm.EnvVars)
	mc, err := dm.ProvisionMachine(machineName, "virtualbox", cmdrunner.RunCommand)
	fmt.Printf("%#v\n", mc)

	if err != nil {
		t.Error(err)
	}

	// cleanup: remove workspacedir/evts/<EVTID>
	_, err = cmdrunner.RunCommand([]string{"VBoxManage", "controlvm", machineName, "poweroff"}, dm.EnvVars)
	assert.Nil(t, err)
	_, err = cmdrunner.RunCommand([]string{"VBoxManage", "unregistervm", machineName}, dm.EnvVars)
	assert.Nil(t, err)
	evtDir, err := utils.Evts(mockSettings, evtID)
	assert.Nil(t, err)
	fmt.Println("Gonna rm -rf", evtDir)
	err = os.RemoveAll(evtDir)
	assert.Nil(t, err)
}
