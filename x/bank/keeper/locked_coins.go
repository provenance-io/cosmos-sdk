package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

var _ types.GetLockedCoinsFn = BaseViewKeeper{}.UnvestedCoins

// UnvestedCoins returns all the coins that are locked due to a vesting schedule.
// It is appended as a GetLockedCoinsFn during NewBaseViewKeeper.
//
// You probably want to call LockedCoins instead. This function is primarily made public
// so that, externally, it can be re-injected after a call to ClearLockedCoinsGetter.
func (k BaseViewKeeper) UnvestedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	if types.HasVestingLockedBypass(ctx) {
		return sdk.NewCoins()
	}
	acc := k.ak.GetAccount(ctx, addr)
	if acc != nil {
		vacc, ok := acc.(types.VestingAccount)
		if ok {
			return vacc.LockedCoins(ctx.BlockTime())
		}
	}
	return sdk.NewCoins()
}
