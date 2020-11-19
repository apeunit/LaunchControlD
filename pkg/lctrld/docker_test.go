package lctrld

import (
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/model"
	"github.com/stretchr/testify/assert"
)

type mockDockerMachineConfig struct {
	WantError bool
}

func (m *mockDockerMachineConfig) HomeDir(machineN string) string {
	n := []string{"mock_path", fmt.Sprint(machineN)}
	return path.Join(n...)
}

func (m *mockDockerMachineConfig) ReadConfig(machineN string) (mc *model.MachineConfig, err error) {
	if m.WantError {
		return nil, errors.New("Here is your new error")
	}

	return new(model.MachineConfig), nil
}

func TestProvision(t *testing.T) {
	fakeGenesisAccounts := []model.GenesisAccount{
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
	evt := model.NewEvent("evtx", "owner", "virtualbox", "nonexistent/testimage", fakeGenesisAccounts, model.LaunchPayload{})

	wantCommandOutput := "1.2.3.4"
	var mockCommandRunner = func(cmd string, args, envVars []string) (out string, err error) {
		fmt.Println("cmd", cmd, "args", args, "envVars", envVars)
		return wantCommandOutput, nil
	}

	dmc := &mockDockerMachineConfig{
		WantError: false,
	}
	evt, err := Provision(settings, evt, mockCommandRunner, dmc)
	assert.Nil(t, err)

	expectedEvt := &model.Event{
		TokenSymbol: "evtx",
		Owner:       "owner",
		Accounts: map[string]*model.Account{
			"first validator": {
				Name:           "first validator",
				Address:        "",
				Mnemonic:       "",
				GenesisBalance: "1000000stake",
				Validator:      true,
				ConfigLocation: &model.ConfigLocation{
					CLIConfigDir:    "",
					DaemonConfigDir: "",
				},
			},
			"second validator": {
				Name:           "second validator",
				Address:        "",
				Mnemonic:       "",
				GenesisBalance: "1000000stake",
				Validator:      true,
				ConfigLocation: &model.ConfigLocation{
					CLIConfigDir:    "",
					DaemonConfigDir: "",
				},
			},
			"third validator": {
				Name:           "third validator",
				Address:        "",
				Mnemonic:       "",
				GenesisBalance: "1000000stake",
				Validator:      true,
				ConfigLocation: &model.ConfigLocation{
					CLIConfigDir:    "",
					DaemonConfigDir: "",
				},
			},
			"hanger on": {
				Name:           "hanger on",
				Address:        "",
				Mnemonic:       "",
				GenesisBalance: "1000000drop,10stake",
				Validator:      false,
				ConfigLocation: &model.ConfigLocation{
					CLIConfigDir:    "",
					DaemonConfigDir: "",
				},
			},
		},
		Provider:    "virtualbox",
		DockerImage: "nonexistent/testimage",
		CreatedOn:   evt.CreatedOn,
		StartsOn:    evt.StartsOn,
		EndsOn:      evt.EndsOn,
		State: map[string]*model.MachineConfig{
			"first validator": {
				N:                "0",
				EventID:          "evtx-2189cd35d97b3f53cc89",
				DriverName:       "",
				TendermintNodeID: "",
				Instance: struct {
					IPAddress   string "json:\"IPAddress\""
					MachineName string "json:\"MachineName\""
					SSHUser     string `json:"SSHUser"`
					SSHPort     int    `json:"SSHPort"`
					SSHKeyPath  string `json:"SSHKeyPath"`
					StorePath   string `json:"StorePath"`
				}{
					IPAddress:   "1.2.3.4",
					MachineName: "",
				},
			},
			"second validator": {
				N:                "1",
				EventID:          "evtx-2189cd35d97b3f53cc89",
				DriverName:       "",
				TendermintNodeID: "",
				Instance: struct {
					IPAddress   string "json:\"IPAddress\""
					MachineName string "json:\"MachineName\""
					SSHUser     string `json:"SSHUser"`
					SSHPort     int    `json:"SSHPort"`
					SSHKeyPath  string `json:"SSHKeyPath"`
					StorePath   string `json:"StorePath"`
				}{
					IPAddress:   "1.2.3.4",
					MachineName: "",
				},
			},
			"third validator": {
				N:                "2",
				EventID:          "evtx-2189cd35d97b3f53cc89",
				DriverName:       "",
				TendermintNodeID: "",
				Instance: struct {
					IPAddress   string "json:\"IPAddress\""
					MachineName string "json:\"MachineName\""
					SSHUser     string `json:"SSHUser"`
					SSHPort     int    `json:"SSHPort"`
					SSHKeyPath  string `json:"SSHKeyPath"`
					StorePath   string `json:"StorePath"`
				}{
					IPAddress:   "1.2.3.4",
					MachineName: "",
				},
			},
		},
	}
	assert.Equal(t, expectedEvt, evt)
}
