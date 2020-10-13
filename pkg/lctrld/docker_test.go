package lctrld

import (
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
)

type mockCommandRunner struct {
	MockOutput string
	WantError  bool
}

func (m *mockCommandRunner) Run(string, []string, []string) (string, error) {
	if m.WantError {
		return "", errors.New("Here is your new error")
	}
	return m.MockOutput, nil
}

type mockDockerMachineConfig struct {
	WantMachineConfig *model.MachineConfig
	WantError         bool
}

func (m *mockDockerMachineConfig) HomeDir(machineN int) string {
	n := []string{"mock_path", fmt.Sprint(machineN)}
	return path.Join(n...)
}

func (m *mockDockerMachineConfig) ReadConfig(machineN int) (mc *model.MachineConfig, err error) {
	if m.WantError {
		return nil, errors.New("Here is your new error")
	}

	return m.WantMachineConfig, nil
}

func TestProvision(t *testing.T) {
	fakeGenesisAccounts := []config.GenesisAccount{
		{
			Name:           "first validator",
			GenesisBalance: "1000000stake",
			Validator:      true,
		},
		{
			Name:           "second validator",
			GenesisBalance: "1000000stake",
			Validator:      true,
		},
		{
			Name:           "third validator",
			GenesisBalance: "1000000stake",
			Validator:      true,
		},
		{
			Name:           "hanger on",
			GenesisBalance: "1000000drop,10stake",
			Validator:      false,
		},
	}
	settings := config.Schema{}
	evt := model.NewEvtvzE("TEST", "owner", "virtualbox", "nonexistent/testimage", fakeGenesisAccounts)
	c := &mockCommandRunner{}
	c.MockOutput = "HELLO WORLD"
	dmc := &mockDockerMachineConfig{
		WantMachineConfig: &model.MachineConfig{
			ID:               "",
			DriverName:       "",
			TendermintNodeID: "",
			Instance: struct {
				IPAddress   string "json:\"IPAddress\""
				MachineName string "json:\"MachineName\""
			}{
				IPAddress:   "",
				MachineName: "",
			},
		},
		WantError: false,
	}
	err := Provision(settings, evt, c, dmc)
	if err != nil {
		t.Fatal(err)
	}
}
