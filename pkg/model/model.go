package model

import (
	"fmt"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/utils"
	"github.com/gosimple/slug"
)

// defaults
// TODO: this will have to be moved somewhere else
const (
	DefaultStake    = 1_000_000
	DefaultCoinbase = 1_000_000_000
	DefaultProvider = "hetzner"
)

// EvtvzE maintain the status of an event
type EvtvzE struct {
	TokenSymbol string                    `json:"token_symbol,omitempty"` // token symbool
	Coinbase    uint64                    `json:"coinbase,omitempty"`     // total amount of tokens at stake
	Stake       uint64                    `json:"stake,omitempty"`        // stake for each validator
	Owner       string                    `json:"owner,omitempty"`        // email address of the owner
	Validators  []string                  `json:"validators,omitempty"`   // email addresses for the validators
	Provider    string                    `json:"provider,omitempty"`     // provider for provisioning
	NodeIPs     []string                  `json:"node_i_ps,omitempty"`    // ip addresses of the nodes
	CreatedOn   time.Time                 `json:"created_on,omitempty"`
	StartsOn    time.Time                 `json:"starts_on,omitempty"`
	EndsOn      time.Time                 `json:"ends_on,omitempty"`
	State       map[string]*MachineConfig `json:"state,omitempty"`
}

// NewEvtvzE helper for a new event
func NewEvtvzE(symbol, owner string, coinbase uint64) (e EvtvzE) {
	return EvtvzE{
		TokenSymbol: symbol,
		Owner:       owner,
		Coinbase:    coinbase,
		Stake:       DefaultStake,
		Provider:    DefaultProvider,
		CreatedOn:   time.Now(),
		StartsOn:    time.Now(),
		State:       make(map[string]*MachineConfig),
	}
}

// ValidatorsCount returns the number of validators
func (e EvtvzE) ValidatorsCount() int {
	return len(e.Validators)
}

// FormatAmount print the amount in a human readable format
func (e EvtvzE) FormatAmount(a uint64) string {
	return fmt.Sprintf("%v%s", a, e.TokenSymbol)
}

// Hash Generate the event hash
func (e EvtvzE) Hash() string {
	return utils.ShortHash(e.TokenSymbol, e.Owner, fmt.Sprint(e.Coinbase))
}

// ID generate a event identifier (determinitstic)
func (e EvtvzE) ID() string {
	return slug.Make(fmt.Sprintf("%v %v", e.TokenSymbol, e.Hash()))
}

// NodeID generate a node identifier (determinitstic)
func (e EvtvzE) NodeID(n int) string {
	return slug.Make(fmt.Sprintf("%v %v %v", e.TokenSymbol, e.Hash(), n))
}

// MachineConfig holds the configuration of a Machine
type MachineConfig struct {
	ID         string `json:"Name,omitempty"`
	DriverName string `json:"driverName,omitempty"`
	Instance   struct {
		IPAddress   string `json:"IPAddress,omitempty"`
		MachineName string `json:"MachineName,omitempty"`
		SSHUser     string `json:"SSHUser,omitempty"`
		SSHPort     int    `json:"SSHPort,omitempty"`
		SSHKeyPath  string `json:"SSHKeyPath,omitempty"`
	} `json:"Driver,omitempty"`
	PayloadConfig PayloadConfig `json:"PayloadConfig,omitempty"`
}

// PayloadConfig stores configuration details for the cosmos-sdk based payload
type PayloadConfig struct {
	Account struct {
		Address  string `json:"address"`
		Mnemonic string `json:"mnemonic"`
	} `json:"Account,omitempty"`
	TendermintNodeID string `json:"TendermintNodeID,omitempty"`
	CLIConfigDir     string `json:"CLIConfigDir,omitempty"`
	DaemonConfigDir  string `json:"DaemonConfigDir,omitempty"`
}
