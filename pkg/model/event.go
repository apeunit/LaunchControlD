package model

import (
	"fmt"
	"sort"
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
	TokenSymbol string              `json:"token_symbol"` // token symbool
	Owner       string              `json:"owner"`        // email address of the owner
	Accounts    map[string]*Account `json:"accounts"`
	Provider    string              `json:"provider"` // provider for provisioning
	CreatedOn   time.Time           `json:"created_on"`
	StartsOn    time.Time           `json:"starts_on"`
	EndsOn      time.Time           `json:"ends_on"`
	State       map[string]*Machine `json:"state"`
	Payload     PayloadLocation     `json:"payload"`
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
		State:       make(map[string]*Machine),
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
	return utils.ShortHash(e.TokenSymbol, e.Owner)
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
