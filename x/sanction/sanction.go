package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultMinDepositSanction is the default to use for the MinDepositSanction.
var DefaultMinDepositSanction sdk.Coins = nil

// DefaultMinDepositUnsanction is the default to use for the MinDepositUnsanction.
var DefaultMinDepositUnsanction sdk.Coins = nil

func DefaultImmediateParams() *ImmediateParams {
	return &ImmediateParams{
		MinDepositSanction:   DefaultMinDepositSanction,
		MinDepositUnsanction: DefaultMinDepositUnsanction,
	}
}

func (p ImmediateParams) ValidateBasic() error {
	if err := p.MinDepositSanction.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrapf("MinDepositSanction: %s", err.Error())
	}
	if err := p.MinDepositUnsanction.Validate(); err != nil {
		return sdkerrors.ErrInvalidCoins.Wrapf("MinDepositUnsanction: %s", err.Error())
	}
	return nil
}
