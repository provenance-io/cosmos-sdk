package quarantine

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Validate performs basic validation of genesis data returning an error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	// TODO[1046]: Implement the GenesisState.Validate function.
	return nil
}

// NewGenesisState creates a new genesis state for the quarantine module.
func NewGenesisState(quarantinedAddresses []string, autoResponses []QuarantineAutoResponseEntry, funds []QuarantinedFunds) *GenesisState {
	return &GenesisState{
		QuarantinedAddresses:    quarantinedAddresses,
		QuarantineAutoResponses: autoResponses,
		QuarantinedFunds:        funds,
	}
}

// DefaultGenesisState returns a default quarantine module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState([]string{}, []QuarantineAutoResponseEntry{}, []QuarantinedFunds{})
}

// GetGenesisStateFromAppState returns x/quarantine GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
