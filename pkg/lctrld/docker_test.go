package lctrld

import (
	"errors"
	"fmt"
	"path"

	"github.com/apeunit/LaunchControlD/pkg/model"
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
