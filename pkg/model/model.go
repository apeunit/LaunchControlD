package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/config"
	"github.com/apeunit/LaunchControlD/pkg/utils"
	"github.com/gosimple/slug"
)

// defaults
// TODO: this will have to be moved somewhere else
const (
	DefaultProvider = "hetzner"
)

// EvtvzE maintain the status of an event
type EvtvzE struct {
	TokenSymbol string                    `json:"token_symbol,omitempty"` // token symbool
	Owner       string                    `json:"owner,omitempty"`        // email address of the owner
	Accounts    map[string]*Account       `json:"accounts,omitempty"`
	Provider    string                    `json:"provider,omitempty"`  // provider for provisioning
	NodeIPs     []string                  `json:"node_i_ps,omitempty"` // ip addresses of the nodes
	CreatedOn   time.Time                 `json:"created_on,omitempty"`
	StartsOn    time.Time                 `json:"starts_on,omitempty"`
	EndsOn      time.Time                 `json:"ends_on,omitempty"`
	State       map[string]*MachineConfig `json:"state,omitempty"`
}

// NewEvtvzE helper for a new event
func NewEvtvzE(symbol, owner, provider string, genesisAccounts []config.GenesisAccount) (e EvtvzE) {
	accounts := make(map[string]*Account)
	for _, acc := range genesisAccounts {
		a := &Account{
			Name:           acc.Name,
			Address:        "",
			Mnemonic:       "",
			GenesisBalance: acc.GenesisBalance,
			Validator:      acc.Validator,
		}

		accounts[acc.Name] = a
	}
	return EvtvzE{
		TokenSymbol: symbol,
		Owner:       owner,
		Accounts:    accounts,
		Provider:    provider,
		NodeIPs:     nil,
		CreatedOn:   time.Now(),
		StartsOn:    time.Now(),
		EndsOn:      time.Time{},
		State:       make(map[string]*MachineConfig),
	}
}

// LoadEvtvzE is a convenience function that ensures you don't have to manually
// create an empty models.Evtvze{} struct and use NewEvtvzE() all the time
func LoadEvtvzE(path string) (evt *EvtvzE, err error) {
	err = utils.LoadJSON(path, &evt)
	return
}

// FormatAmount print the amount in a human readable format
func (e EvtvzE) FormatAmount(a uint64) string {
	return fmt.Sprintf("%v%s", a, e.TokenSymbol)
}

// Hash Generate the event hash
func (e EvtvzE) Hash() string {
	return utils.ShortHash(e.TokenSymbol, e.Owner, e.Provider)
}

// ID generate a event identifier (determinitstic)
func (e EvtvzE) ID() string {
	return slug.Make(fmt.Sprintf("%v %v", e.TokenSymbol, e.Hash()))
}

// NodeID generate a node identifier (determinitstic)
func (e EvtvzE) NodeID(n int) string {
	return slug.Make(fmt.Sprintf("%v %v %v", e.TokenSymbol, e.Hash(), n))
}

func (e EvtvzE) sortedAccounts() (keys []string) {
	// We need to return the accounts in a deterministic order, sorted by key
	for k := range e.Accounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Validators returns the names (emails) of the validators
func (e EvtvzE) Validators() (v []string, a []*Account) {
	for _, k := range e.sortedAccounts() {
		if e.Accounts[k].Validator {
			v = append(v, e.Accounts[k].Name)
			a = append(a, e.Accounts[k])
		}
	}
	return
}

// ValidatorsCount returns the number of validators
func (e EvtvzE) ValidatorsCount() int {
	validatorNames, _ := e.Validators()
	return len(validatorNames)
}

// ExtraAccounts returns EvtvzE.Accounts excluding accounts that are validators
func (e EvtvzE) ExtraAccounts() (a []*Account) {
	for _, k := range e.sortedAccounts() {
		if !e.Accounts[k].Validator {
			a = append(a, e.Accounts[k])
		}
	}
	return
}

// MachineConfig holds the configuration of a Machine
type MachineConfig struct {
	ID               string `json:"Name,omitempty"`
	DriverName       string `json:"DriverName,omitempty"`
	TendermintNodeID string `json:"TendermintNodeID,omitempty"`
	CLIConfigDir     string `json:"CLIConfigDir,omitempty"`
	DaemonConfigDir  string `json:"DaemonConfigDir,omitempty"`
	Instance         struct {
		IPAddress   string `json:"IPAddress,omitempty"`
		MachineName string `json:"MachineName,omitempty"`
		SSHUser     string `json:"SSHUser,omitempty"`
		SSHPort     int    `json:"SSHPort,omitempty"`
		SSHKeyPath  string `json:"SSHKeyPath,omitempty"`
	} `json:"Instance,omitempty"`
}

func (m MachineConfig) NumberID() (numberID int, err error) {
	splitID := strings.Split(m.ID, "-") // evtx-d97517a3673688070aef-1
	nID, err := strconv.ParseInt(splitID[len(splitID)-1], 0, 64)
	return int(nID), err
}

// TendermintPeerNodeID returns <nodeID@192.168.1....:26656> which is used in specifying peers to connect to in the daemon's config.toml file
func (m MachineConfig) TendermintPeerNodeID() string {
	return fmt.Sprintf("%s@%s:26656", m.TendermintNodeID, m.Instance.IPAddress)
}

type Account struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	Mnemonic       string `json:"mnemonic"`
	GenesisBalance string `json:"genesis_balance"`
	Validator      bool   `json:"validator"`
}

func NewAccount(name, address, mnemonic, genesisBalance string, validator bool) *Account {
	return &Account{
		Name:           name,
		Address:        address,
		Mnemonic:       mnemonic,
		GenesisBalance: genesisBalance,
		Validator:      validator,
	}
}
