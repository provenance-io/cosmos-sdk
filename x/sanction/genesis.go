package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

func NewGenesisState(params *Params, addrs []string) *GenesisState {
	return &GenesisState{
		Params:              params,
		SanctionedAddresses: addrs,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), nil)
}

func (g GenesisState) Validate() error {
	if g.Params != nil {
		if err := g.Params.ValidateBasic(); err != nil {
			return errors.ErrInvalidParams.Wrap(err.Error())
		}
	}
	for i, addr := range g.SanctionedAddresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("sanctioned addresses[%d], %q: %v", i, addr, err)
		}
	}
	return nil
}
