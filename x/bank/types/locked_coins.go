package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const bypassKey = "bypass-vesting-locked-coins" //nolint:gosec // Not actually credentials.

// WithVestingLockedBypass returns a new context that will cause the vesting locked coins lookup to be skipped.
func WithVestingLockedBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, true)
	return context.Context(sdkCtx).(C)
}

// WithoutVestingLockedBypass returns a new context that will cause the vesting locked coins lookup to not be skipped.
func WithoutVestingLockedBypass[C context.Context](ctx C) C {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithValue(bypassKey, false)
	return context.Context(sdkCtx).(C)
}

// HasVestingLockedBypass checks the context to see if the vesting locked coins lookup should be skipped.
func HasVestingLockedBypass[C context.Context](ctx C) bool {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	bypassValue := sdkCtx.Value(bypassKey)
	if bypassValue == nil {
		return false
	}
	bypass, isBool := bypassValue.(bool)
	return isBool && bypass
}

// A GetLockedCoinsFn returns some coins locked for an address.
type GetLockedCoinsFn func(ctx context.Context, addr sdk.AccAddress) sdk.Coins

var _ GetLockedCoinsFn = NoOpGetLockedCoinsFn

// NoOpGetLockedCoinsFn is a no-op GetLockedCoinsFn.
func NoOpGetLockedCoinsFn(_ context.Context, _ sdk.AccAddress) sdk.Coins {
	return sdk.NewCoins()
}

// Then creates a composite restriction that runs this one then the provided second one.
func (r GetLockedCoinsFn) Then(second GetLockedCoinsFn) GetLockedCoinsFn {
	return ComposeGetLockedCoins(r, second)
}

// ComposeGetLockedCoins combines multiple GetLockedCoinsFn into one.
// nil entries are ignored.
// If all entries are nil, nil is returned.
// If exactly one entry is not nil, it is returned.
// Otherwise, a new GetLockedCoinsFn is returned that runs the non-nil functions in the order they are given.
// The composition runs each function returning the sum of all results.
func ComposeGetLockedCoins(restrictions ...GetLockedCoinsFn) GetLockedCoinsFn {
	toRun := make([]GetLockedCoinsFn, 0, len(restrictions))
	for _, r := range restrictions {
		if r != nil {
			toRun = append(toRun, r)
		}
	}
	switch len(toRun) {
	case 0:
		return nil
	case 1:
		return toRun[0]
	}
	return func(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
		rv := sdk.NewCoins()
		for _, f := range toRun {
			newLocked := f(ctx, addr)
			if !newLocked.IsZero() {
				rv = rv.Add(newLocked...)
			}
		}
		return rv
	}
}
