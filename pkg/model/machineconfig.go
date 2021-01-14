package model

import (
	"fmt"
	"strings"
)

// MachineConfig holds the configuration of a Machine
type MachineConfig struct {
	N                string                `json:"N"`
	EventID          string                `json:"EventID"`
	DriverName       string                `json:"DriverName"`
	TendermintNodeID string                `json:"TendermintNodeID"`
	Instance         MachineConfigInstance `json:"Instance"`
}

// ID joins the EventID and N, e.g. EventID is evtx-d97517a3673688070aef, N is
// 1, then it will return evtx-d97517a3673688070aef-1
func (m *MachineConfig) ID() string {
	s := []string{m.EventID, m.N}
	return strings.Join(s, "-")
}

// TendermintPeerNodeID returns <nodeID@192.168.1....:26656> which is used in specifying peers to connect to in the daemon's config.toml file
func (m *MachineConfig) TendermintPeerNodeID() string {
	return fmt.Sprintf("%s@%s:26656", m.TendermintNodeID, m.Instance.IPAddress)
}

// MachineConfigInstance holds information read from docker-machine about the deployed VM's network settings
type MachineConfigInstance struct {
	IPAddress   string `json:"IPAddress"`
	MachineName string `json:"MachineName"`
	SSHUser     string `json:"SSHUser"`
	SSHPort     int    `json:"SSHPort"`
	SSHKeyPath  string `json:"SSHKeyPath"`
	StorePath   string `json:"StorePath"`
}
