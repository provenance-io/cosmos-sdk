package quarantine

import (
	"cosmossdk.io/errors"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Validate performs basic validation of genesis data returning an error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	for _, addr := range gs.QuarantinedAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid quarantined address: %v", err)
		}
	}
	for _, resp := range gs.AutoResponses {
		if err := resp.Validate(); err != nil {
			return errors.Wrap(err, "invalid quarantine auto response entry")
		}
	}
	for _, funds := range gs.QuarantinedFunds {
		if err := funds.Validate(); err != nil {
			return errors.Wrap(err, "invalid quarantined funds")
		}
	}
	return nil
}

// NewGenesisState creates a new genesis state for the quarantine module.
func NewGenesisState(quarantinedAddresses []string, autoResponses []*AutoResponseEntry, funds []*QuarantinedFunds) *GenesisState {
	return &GenesisState{
		QuarantinedAddresses: quarantinedAddresses,
		AutoResponses:        autoResponses,
		QuarantinedFunds:     funds,
	}
}

// DefaultGenesisState returns a default quarantine module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil, nil, nil)
}

// GetGenesisStateFromAppState returns x/quarantine GenesisState given raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}
