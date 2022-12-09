package sanction

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewGenesisState(params *ImmediateParams, addrs []string) *GenesisState {
	return &GenesisState{
		ImmediateParams:     params,
		SanctionedAddresses: addrs,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultImmediateParams(), nil)
}

func (g GenesisState) Validate() error {
	if g.ImmediateParams != nil {
		if err := g.ImmediateParams.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "invalid immediate params")
		}
	}
	for i, addr := range g.SanctionedAddresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("sanctioned addresses[%d]", i)
		}
	}
	return nil
}
