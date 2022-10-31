package cli

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	// exampleAddr1 is a random address to use in examples.
	exampleAddr1 = authtypes.NewModuleAddress("exampleAddr1")
	// exampleAddr2 is another random address to use in examples.
	exampleAddr2 = authtypes.NewModuleAddress("exampleAddr2")
)

// validateAddress checks to make sure the provided addr is a valid Bech32 address string.
// If it is invalid, "" is returned with an error that includes the argName.
// If it is valid, the addr is returned without an error.
//
// This validation is (hopefully) already done by the node, but it's more
// user-friendly to also do it here, before a request is actually sent.
func validateAddress(addr string, argName string) (string, error) {
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return "", sdkerrors.ErrInvalidAddress.Wrapf("invalid %s: %v", argName, err)
	}
	return addr, nil
}
