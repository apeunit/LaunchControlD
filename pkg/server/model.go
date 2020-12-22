package server

import (
	"net/http"
	"time"

	"github.com/apeunit/LaunchControlD/pkg/model"
)

// UserCredentials the input user credential for authentication
type UserCredentials struct {
	Email string `json:"email,omitempty"`
	Pass  string `json:"pass,omitempty"`
}

// APIReply a reply from the API
type APIReply struct {
	Status  int    `json:"code"`
	Message string `json:"message"`
}

// APIStatus hold the status of the API
type APIStatus struct {
	Status  string `json:"status,omitempty"`
	Version string `json:"version,omitempty"`
	Uptime  string `json:"uptime,omitempty"`
}

// APIReplyOK returns an 200 reply
func APIReplyOK(m string) APIReply {
	return APIReply{
		Status:  http.StatusOK,
		Message: m,
	}
}

// APIReplyErr error reply
func APIReplyErr(code int, m string) APIReply {
	return APIReply{
		Status:  code,
		Message: m,
	}
}

// APIEvent API safe event object
type APIEvent struct {
	ID          string                      `json:"id"`
	TokenSymbol string                      `json:"token_symbol"` // token symbool
	Owner       string                      `json:"owner"`        // email address of the owner
	Accounts    map[string]APIAccount       `json:"accounts"`
	Provider    string                      `json:"provider"` // provider for provisioning
	CreatedOn   time.Time                   `json:"created_on"`
	StartsOn    time.Time                   `json:"starts_on"`
	EndsOn      time.Time                   `json:"ends_on"`
	State       map[string]APIMachineConfig `json:"state"`
}

// APIAccount API safe account object
type APIAccount struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	GenesisBalance string `json:"genesis_balance"`
	Validator      bool   `json:"validator"`
	Faucet         bool   `json:"faucet"`
}

// APIMachineConfig API safe machine config
type APIMachineConfig struct {
	TendermintNodeID string `json:"tendermint_node_id"`
	IPAddress        string `json:"IPAddress"`
	MachineName      string `json:"MachineName"`
}

// ToAPIEvents copy a list of events to a API save version
func ToAPIEvents(evts *[]model.Event) (aEvts []APIEvent) {
	aEvts = make([]APIEvent, len(*evts))
	for _, v := range *evts {
		aEvts = append(aEvts, ToAPIEvent(&v))
	}
	return
}

// ToAPIEvent convert and Event to an APIEvent that is
// safe to publish via REST API Endpoints
func ToAPIEvent(evt *model.Event) (aEvt APIEvent) {
	aEvt = APIEvent{
		ID:          evt.ID(),
		TokenSymbol: evt.TokenSymbol,
		Owner:       evt.Owner,
		Provider:    evt.Provider,
		CreatedOn:   evt.CreatedOn,
		StartsOn:    evt.StartsOn,
		EndsOn:      evt.EndsOn,
		Accounts:    make(map[string]APIAccount, len(evt.Accounts)),
		State:       make(map[string]APIMachineConfig, len(evt.State)),
	}
	// fill up the accounts
	for k, v := range evt.Accounts {
		aEvt.Accounts[k] = APIAccount{
			Name:           v.Name,
			Address:        v.Address,
			GenesisBalance: v.GenesisBalance,
			Validator:      v.Validator,
			Faucet:         v.Faucet,
		}
	}
	// fill up the machine config
	for k, v := range evt.State {
		aEvt.State[k] = APIMachineConfig{
			TendermintNodeID: v.TendermintNodeID,
			IPAddress:        v.Instance.IPAddress,
			MachineName:      v.Instance.MachineName,
		}
	}
	return
}
