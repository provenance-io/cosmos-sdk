package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ Authorization = &CountAuthorization{}
)

// NewCountAuthorization creates a new CountAuthorization object.
func NewCountAuthorization(msgTypeURL string, remainingAuthorizations int32) *CountAuthorization {
	return &CountAuthorization{
		Msg: msgTypeURL,
		AllowedAuthorizations: remainingAuthorizations,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a CountAuthorization) MsgTypeURL() string {
	return a.Msg
}

// Accept implements Authorization.Accept.
func (a CountAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (AcceptResponse, error) {
	remaining := a.AllowedAuthorizations
	isNegative := remaining < 0
	if isNegative {
		return AcceptResponse{}, sdkerrors.ErrInvalidRequest.Wrapf("allowed authorizations cannot be negative")
	}
	if remaining == 0 {
		return AcceptResponse{Accept: true, Delete: true}, nil
	}

	return AcceptResponse{Accept: true, Delete: false, Updated: &CountAuthorization{Msg: a.Msg, AllowedAuthorizations: a.AllowedAuthorizations}}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a CountAuthorization) ValidateBasic() error {
	if a.AllowedAuthorizations < 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf("allowed authorizations cannot be negative")
	}
	return nil
}
