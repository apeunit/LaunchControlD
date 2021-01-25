package model

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apeunit/LaunchControlD/pkg/cmdrunner"
	"github.com/apeunit/LaunchControlD/pkg/config"
)

// Machine holds the configuration of a Machine
type Machine struct {
	N                string               `json:"N"`
	EventID          string               `json:"EventID"`
	DriverName       string               `json:"DriverName"`
	TendermintNodeID string               `json:"TendermintNodeID"`
	Instance         MachineNetworkConfig `json:"Instance"`
	settings         config.Schema
	dockerMachineEnv []string
	CmdRunner        cmdrunner.CommandRunner
}

// ID joins the EventID and N, e.g. EventID is evtx-d97517a3673688070aef, N is
// 1, then it will return evtx-d97517a3673688070aef-1
func (m *Machine) ID() string {
	s := []string{m.EventID, m.N}
	return strings.Join(s, "-")
}

// HomeDir returns the path of a docker-machine instance home, e.g.
// /tmp/workspace/evts/drop-xxx/.docker/machine/machines/drop-xxx-0/
// TODO: reconcile with duplicate logic from lctrld/common.go
func (m *Machine) HomeDir() string {
	return filepath.Join(m.settings.Workspace, "evts", m.EventID, ".docker", "machine", "machines", fmt.Sprintf("%s-%s", m.EventID, m.N))
}

// TendermintPeerNodeID returns <nodeID@192.168.1....:26656> which is used in specifying peers to connect to in the daemon's config.toml file
func (m *Machine) TendermintPeerNodeID() string {
	return fmt.Sprintf("%s@%s:26656", m.TendermintNodeID, m.Instance.IPAddress)
}

// MachineNetworkConfig holds information read from docker-machine about the deployed VM's network settings
type MachineNetworkConfig struct {
	IPAddress   string `json:"IPAddress"`
	MachineName string `json:"MachineName"`
	SSHUser     string `json:"SSHUser"`
	SSHPort     int    `json:"SSHPort"`
	SSHKeyPath  string `json:"SSHKeyPath"`
	StorePath   string `json:"StorePath"`
}
