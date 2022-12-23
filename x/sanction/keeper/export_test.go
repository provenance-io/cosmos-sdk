package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

var (
	// ConcatBzPlusCap, for unit tests, exposes the concatBzPlusCap function.
	ConcatBzPlusCap = concatBzPlusCap

	// ToCoinsOrDefault, for unit tests, exposes the toCoinsOrDefault function.
	ToCoinsOrDefault = toCoinsOrDefault

	// ToAccAddrs, for unit tests, exposes the toAccAddrs function.
	ToAccAddrs = toAccAddrs
)

// WithGovKeeper, for unit tests, creates a copy of this, setting the govKeeper to the provided one.
func (k Keeper) WithGovKeeper(govKeeper sanction.GovKeeper) Keeper {
	k.govKeeper = govKeeper
	return k
}

// WithAuthority, for unit tests, creates a copy of this, setting the authority to the provided one.
func (k Keeper) WithAuthority(authority string) Keeper {
	k.authority = authority
	return k
}

// WithUnsanctionableAddrs, for unit tests, creates a copy of this, setting the unsanctionableAddrs to the provided one.
// This does not add the provided ones to the unsanctionableAddrs, it overwrites the
// existing ones with the ones provided.
func (k Keeper) WithUnsanctionableAddrs(unsanctionableAddrs map[string]bool) Keeper {
	k.unsanctionableAddrs = unsanctionableAddrs
	return k
}

// GetCodec, for unit tests, exposes this keeper's codec (cdc).
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// GetStoreKey, for unit tests, exposes this keeper's storekey.
func (k Keeper) GetStoreKey() storetypes.StoreKey {
	return k.storeKey
}

// GetMsgSanctionTypeURL, for unit tests, exposes this keeper's msgSanctionTypeURL.
func (k Keeper) GetMsgSanctionTypeURL() string {
	return k.msgSanctionTypeURL
}

// GetMsgSanctionTypeURL, for unit tests, exposes this keeper's msgUnsanctionTypeURL.
func (k Keeper) GetMsgUnsanctionTypeURL() string {
	return k.msgUnsanctionTypeURL
}

// GetMsgSanctionTypeURL, for unit tests, exposes this keeper's msgExecLegacyContentTypeURL.
func (k Keeper) GetMsgExecLegacyContentTypeURL() string {
	return k.msgExecLegacyContentTypeURL
}

// GetParamAsCoinsOrDefault, for unit tests, exposes this keeper's getParamAsCoinsOrDefault function.
func (k Keeper) GetParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	return k.getParamAsCoinsOrDefault(ctx, name, dflt)
}

// GetLatestTempEntry, for unit tests, exposes this keeper's getLatestTempEntry function.
func (k Keeper) GetLatestTempEntry(store sdk.KVStore, addr sdk.AccAddress) []byte {
	return k.getLatestTempEntry(store, addr)
}

// GetParam, for unit tests, exposes this keeper's getParam function.
func (k Keeper) GetParam(store sdk.KVStore, name string) (string, bool) {
	return k.getParam(store, name)
}

// SetParam, for unit tests, exposes this keeper's setParam function.
func (k Keeper) SetParam(store sdk.KVStore, name, value string) {
	k.setParam(store, name, value)
}

// DeleteParam, for unit tests, exposes this keeper's deleteParam function.
func (k Keeper) DeleteParam(store sdk.KVStore, name string) {
	k.deleteParam(store, name)
}
