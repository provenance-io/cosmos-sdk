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
