package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// WithFundsHolder creates a copy of this setting the funds holder to the provided addr.
func (k Keeper) WithFundsHolder(addr sdk.AccAddress) Keeper {
	k.fundsHolder = addr
	return k
}
