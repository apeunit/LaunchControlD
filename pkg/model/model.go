package model

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/utils"
	"github.com/gosimple/slug"
)

// defaults
// TODO: this will have to be moved somewhere else
const (
	DefaultProvider = "hetzner"
)

// Event maintain the status of an event
type Event struct {
	TokenSymbol string                    `json:"token_symbol"` // token symbool
	Owner       string                    `json:"owner"`        // email address of the owner
	Accounts    map[string]*Account       `json:"accounts"`
	Provider    string                    `json:"provider"` // provider for provisioning
	CreatedOn   time.Time                 `json:"created_on"`
	StartsOn    time.Time                 `json:"starts_on"`
	EndsOn      time.Time                 `json:"ends_on"`
	State       map[string]*MachineConfig `json:"state"`
	Payload     PayloadLocation           `json:"payload"`
}

// NewEvent helper for a new event
func NewEvent(symbol, owner, provider string, genesisAccounts []GenesisAccount, payload PayloadLocation) (e *Event) {
	accounts := make(map[string]*Account)
	for _, acc := range genesisAccounts {
		accounts[acc.Name] = &Account{
			Name:           acc.Name,
			Address:        "",
			Mnemonic:       "",
			GenesisBalance: acc.GenesisBalance,
			Validator:      acc.Validator,
			Faucet:         acc.Faucet,
			ConfigLocation: ConfigLocation{
				CLIConfigDir:    "",
				DaemonConfigDir: "",
			},
		}
	}
	return &Event{
		TokenSymbol: symbol,
		Owner:       owner,
		Accounts:    accounts,
		Provider:    provider,
		CreatedOn:   time.Now(),
		StartsOn:    time.Now(),
		EndsOn:      time.Time{},
		State:       make(map[string]*MachineConfig),
		Payload:     payload,
	}
}

// LoadEvent is a convenience function that ensures you don't have to manually
// create an empty models.Event{} struct and use NewEvent() all the time
func LoadEvent(path string) (evt *Event, err error) {
	err = utils.LoadJSON(path, &evt)
	return
}

// FormatAmount print the amount in a human readable format
func (e *Event) FormatAmount(a uint64) string {
	return fmt.Sprintf("%v%s", a, e.TokenSymbol)
}

// Hash Generate the event hash
func (e *Event) Hash() string {
	return utils.ShortHash(e.TokenSymbol, e.Owner, e.Provider)
}

// ID generate a event identifier (determinitstic)
func (e *Event) ID() string {
	return slug.Make(fmt.Sprintf("%v %v", e.TokenSymbol, e.Hash()))
}

// NodeID generate a node identifier (determinitstic)
func (e *Event) NodeID(n int) string {
	return slug.Make(fmt.Sprintf("%v %v %v", e.TokenSymbol, e.Hash(), n))
}

func (e *Event) sortedAccounts() (keys []string) {
	// We need to return the accounts in a deterministic order, sorted by key
	for k := range e.Accounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Validators returns the names (emails) of the validators
func (e *Event) Validators() (v []string, a []*Account) {
	for _, k := range e.sortedAccounts() {
		if e.Accounts[k].Validator {
			v = append(v, e.Accounts[k].Name)
			a = append(a, e.Accounts[k])
		}
	}
	return
}

// ValidatorsCount returns the number of validators
func (e *Event) ValidatorsCount() int {
	validatorNames, _ := e.Validators()
	return len(validatorNames)
}

// ExtraAccounts returns Event.Accounts excluding accounts that are validators
func (e *Event) ExtraAccounts() (a []*Account) {
	for _, k := range e.sortedAccounts() {
		if !e.Accounts[k].Validator {
			a = append(a, e.Accounts[k])
		}
	}
	return
}

// FaucetAccount returns the first Account found with Faucet = true
func (e *Event) FaucetAccount() (a *Account) {
	for _, k := range e.sortedAccounts() {
		if e.Accounts[k].Faucet {
			return e.Accounts[k]
		}
	}
	return nil
}

// ConfigLocation holds the paths to the configuration files for the Cosmos-SDK
// based node and CLI.
type ConfigLocation struct {
	CLIConfigDir    string `json:"CLIConfigDir"`
	DaemonConfigDir string `json:"DaemonConfigDir"`
}

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

// Account represents an Account for a Event
type Account struct {
	Name           string         `json:"name"`
	Address        string         `json:"address"`
	Mnemonic       string         `json:"mnemonic"`
	GenesisBalance string         `json:"genesis_balance"`
	Validator      bool           `json:"validator"`
	Faucet         bool           `json:"faucet"`
	ConfigLocation ConfigLocation `json:"config_location"`
}
