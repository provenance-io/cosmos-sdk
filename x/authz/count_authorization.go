package authz

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_         Authorization = &CountAuthorization{}
	errMsgGt0               = "allowed authorizations must be greater than 0"
)

// NewCountAuthorization creates a new CountAuthorization object.
func NewCountAuthorization(msgTypeURL string, allowedAuthorizations int32) *CountAuthorization {
	return &CountAuthorization{
		Msg:                   msgTypeURL,
		AllowedAuthorizations: allowedAuthorizations,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a CountAuthorization) MsgTypeURL() string {
	return a.Msg
}

// Accept implements Authorization.Accept.
func (a CountAuthorization) Accept(_ context.Context, _ sdk.Msg) (AcceptResponse, error) {
	if a.AllowedAuthorizations <= 0 {
		return AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf(errMsgGt0)
	}

	resp := AcceptResponse{Accept: true}
	if a.AllowedAuthorizations == 1 {
		resp.Delete = true
	} else {
		resp.Updated = &CountAuthorization{Msg: a.Msg, AllowedAuthorizations: a.AllowedAuthorizations - 1}
	}

	return resp, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a CountAuthorization) ValidateBasic() error {
	if a.AllowedAuthorizations <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrapf(errMsgGt0)
	}
	return nil
}
