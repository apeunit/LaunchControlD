package lctrld

import (
	"testing"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestDockerMachineConfig(t *testing.T) {
	settings := config.Schema{
		Workspace: "./testdata",
		DockerMachine: config.DockerMachine{
			Workspace:  "",
			SearchPath: nil,
			Version:    "",
			BinaryURL:  "",
			Binary:     "",
			Drivers: map[string]config.DockerMachineDriver{
				"": {
					Version:   "",
					BinaryURL: "",
					Binary:    "",
					Params:    nil,
					Env:       nil,
				},
			},
		},
	}
	dmc := NewDockerMachine(settings, "drop-28b10d4eff415a7b0b2c")
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
