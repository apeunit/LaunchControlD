package lctrld

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/stretchr/testify/assert"
)

var mockSettings = &config.Schema{
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

func TestDockerMachineProvisionMachine(t *testing.T) {
	evtID := "test-startmachine"
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
	evtDir, err := mockSettings.Evts(evtID)
	assert.Nil(t, err)
	fmt.Println("Gonna rm -rf", evtDir)
	err = os.RemoveAll(evtDir)
	assert.Nil(t, err)
}

func TestDockerMachineStopMachine(t *testing.T) {
	evtID := "test-stopmachine"
	machineName := fmt.Sprintf("%s-%s", evtID, "0")
	dm := NewDockerMachine(mockSettings, evtID)
	fmt.Println(dm.EnvVars)
	_, err := dm.ProvisionMachine(machineName, "virtualbox", cmdrunner.RunCommand)

	if err != nil {
		t.Error(err)
	}

	fmt.Println("Gonna stop the machine")
	err = dm.StopMachine(machineName, cmdrunner.RunCommand)
	if err != nil {
		t.Error(err)
	}
	// cleanup: remove workspacedir/evts/<EVTID>
	evtDir, err := mockSettings.Evts(evtID)
	assert.Nil(t, err)
	fmt.Println("Gonna rm -rf", evtDir)
	err = os.RemoveAll(evtDir)
	assert.Nil(t, err)
}

func TestDockerMachineGeneral(t *testing.T) {
	var err error // Declare it first, in case I want to comment out code
	evtID := "test-general"
	machineName := fmt.Sprintf("%s-%s", evtID, "0")
	dm := NewDockerMachine(mockSettings, evtID)
	fmt.Println(dm.EnvVars)
	_, err = dm.ProvisionMachine(machineName, "virtualbox", cmdrunner.RunCommand)

	if err != nil {
		t.Error(err)
	}

	t.Run("testRun", func(t *testing.T) {
		out, err := dm.Run(machineName, []string{"hostname"}, cmdrunner.RunCommand)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, out, machineName)
		_, err = dm.Run(machineName, []string{"mkdir", "/home/docker/testdir"}, cmdrunner.RunCommand)
		if err != nil {
			t.Error(err)
		}
		out, err = dm.Run(machineName, []string{"ls", "-la", "/home/docker/"}, cmdrunner.RunCommand)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("out", out)
		assert.Contains(t, out, "testdir")
	})
	t.Run("testCopy", func(t *testing.T) {
		evtDir, err := mockSettings.Evts(evtID)
		assert.Nil(t, err)
		testDir := filepath.Join(evtDir, "ThisIsATestDir")
		err = os.Mkdir(testDir, 0755)
		assert.Nil(t, err)

		err = dm.Copy(machineName, testDir, "/home/docker", cmdrunner.RunCommand)
		assert.Nil(t, err)

		out, err := dm.Run(machineName, []string{"ls", "-la", "/home/docker/"}, cmdrunner.RunCommand)
		if err != nil {
			t.Error(err)
		}
		fmt.Println("out", out)
		assert.Contains(t, out, "ThisIsATestDir")
	})

	fmt.Println("Gonna stop the machine")
	err = dm.StopMachine(machineName, cmdrunner.RunCommand)
	if err != nil {
		t.Error(err)
	}
	// cleanup: remove workspacedir/evts/<EVTID>
	evtDir, err := mockSettings.Evts(evtID)
	assert.Nil(t, err)
	fmt.Println("Gonna rm -rf", evtDir)
	err = os.RemoveAll(evtDir)
	assert.Nil(t, err)
}
