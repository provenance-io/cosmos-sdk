package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// SetFundsHolder is only for unit tests and sets the fundsHolder on the keeper.
func (k *Keeper) SetFundsHolder(addr sdk.AccAddress) {
	k.fundsHolder = addr
}
