package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper sanction.BankKeeper

	authority string

	unsanctionableAddrs map[string]bool
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper sanction.BankKeeper, authority string, unsanctionableAddrs []sdk.AccAddress) Keeper {
	rv := Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		bankKeeper:          bankKeeper,
		authority:           authority,
		unsanctionableAddrs: make(map[string]bool),
	}
	for _, addr := range unsanctionableAddrs {
		rv.unsanctionableAddrs[addr.String()] = true
	}
	bankKeeper.SetSanctionKeeper(rv)
	return rv
}

// GetAuthority returns this module's authority string.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsSanctionedAddr -returns true if the provided address is currently sanctioned.
func (k Keeper) IsSanctionedAddr(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	tempEntry := k.getLatestTempEntry(store, addr)
	if IsTempSanctionBz(tempEntry) {
		return true
	}
	if IsTempUnsanctionBz(tempEntry) {
		return false
	}
	key := CreateSanctionedAddrKey(addr)
	return store.Has(key)
}

// SanctionAddresses creates sanctioned address entries for each of the provided addresses.
func (k Keeper) SanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	val := []byte{0x00}
	for _, addr := range addrs {
		key := CreateSanctionedAddrKey(addr)
		store.Set(key, val)
	}
}

// UnsanctionAddresses deletes any sanctioned address entries for each provided address.
func (k Keeper) UnsanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	for _, addr := range addrs {
		key := CreateSanctionedAddrKey(addr)
		store.Delete(key)
	}
}

// AddTemporarySanction adds a temporary sanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporarySanction(ctx sdk.Context, govPropId uint64, addrs ...sdk.AccAddress) {
	k.addTempEntries(ctx, TempSanctionB, govPropId, addrs)
}

// AddTemporaryUnsanction adds a temporary unsanction with the given gov prop id for each of the provided addresses.
func (k Keeper) AddTemporaryUnsanction(ctx sdk.Context, govPropId uint64, addrs ...sdk.AccAddress) {
	k.addTempEntries(ctx, TempUnsanctionB, govPropId, addrs)
}

// addTempEntries adds a temporary entry with the given value and gov prop id for each address given.
func (k Keeper) addTempEntries(ctx sdk.Context, value byte, govPropId uint64, addrs []sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	val := []byte{value}
	for _, addr := range addrs {
		key := CreateTemporaryKey(addr, govPropId)
		store.Set(key, val)
	}
}

// getLatestTempEntry gets the most recent temporary entry for the given address.
func (k Keeper) getLatestTempEntry(store sdk.KVStore, addr sdk.AccAddress) []byte {
	pre := CreateTemporaryAddrPrefix(addr)
	preStore := prefix.NewStore(store, pre)
	iter := preStore.ReverseIterator(nil, nil)
	defer iter.Close()
	if iter.Valid() {
		return iter.Value()
	}
	return nil
}

// DeleteSpecificTempEntries deletes the temporary entries with the given addresses for a specific governance proposal id.
func (k Keeper) DeleteSpecificTempEntries(ctx sdk.Context, govPropId uint64, addrs ...sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	for _, addr := range addrs {
		key := CreateTemporaryKey(addr, govPropId)
		store.Delete(key)
	}
}

// DeleteTempEntries deletes all temporary entries for each given address.
func (k Keeper) DeleteTempEntries(ctx sdk.Context, addrs ...sdk.AccAddress) {
	if len(addrs) == 0 {
		return
	}
	var toRemove [][]byte
	callback := func(cbAddr sdk.AccAddress, govPropId uint64, _ bool) bool {
		toRemove = append(toRemove, CreateTemporaryKey(cbAddr, govPropId))
		return false
	}
	for _, addr := range addrs {
		if len(addr) > 0 {
			k.IterateTemporaryEntries(ctx, addr, callback)
		}
	}
	store := ctx.KVStore(k.storeKey)
	for _, key := range toRemove {
		store.Delete(key)
	}
}

// getSanctionedAddressPrefixStore returns a kv store prefixed for sanctioned addresses, and the prefix bytes.
func (k Keeper) getSanctionedAddressPrefixStore(ctx sdk.Context) (sdk.KVStore, []byte) {
	return prefix.NewStore(ctx.KVStore(k.storeKey), SanctionedPrefix), SanctionedPrefix
}

// IterateSanctionedAddresses iterates over all of the permanently sanctioned addresses.
// The callback takes in the sanctioned address and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateSanctionedAddresses(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool)) {
	store, _ := k.getSanctionedAddressPrefixStore(ctx)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr, _ := ParseLengthPrefixedBz(iter.Key())
		if cb(addr) {
			break
		}
	}
}

// getTemporaryEntryPrefixStore returns a kv store prefixed for temporary sanction/unsanction entries, and the prefix bytes used.
// If an addr is provided, the store is prefixed for just the given address.
// If addr is empty, it will be prefixed for all temporary entries.
func (k Keeper) getTemporaryEntryPrefixStore(ctx sdk.Context, addr sdk.AccAddress) (sdk.KVStore, []byte) {
	pre := CreateTemporaryAddrPrefix(addr)
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre), pre
}

// IterateTemporaryEntries iterates over each of the temporary entries.
// If an address is provided, only the temporary entries for that address are iterated,
// otherwise all entries are iterated.
// The callback takes in the address in question, the governance proposal associated with it, and whether it's a sanction (true) or unsanction (false).
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateTemporaryEntries(ctx sdk.Context, addr sdk.AccAddress, cb func(addr sdk.AccAddress, govPropId uint64, isSanction bool) (stop bool)) {
	store, pre := k.getTemporaryEntryPrefixStore(ctx, addr)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key := ConcatBz(pre, iter.Key())
		kAddr, govPropId := ParseTemporaryKey(key)
		isSanction := IsTempSanctionBz(iter.Value())
		if cb(kAddr, govPropId, isSanction) {
			break
		}
	}
}

// IsSanctionableAddr returns true if the provided address is not one of the ones that cannot be sanctioned.
// I.e. returns true if it can be sanctioned.
func (k Keeper) IsSanctionableAddr(addr string) bool {
	return !k.unsanctionableAddrs[addr]
}

// GetParams gets the sanction module's params.
func (k Keeper) GetParams(ctx sdk.Context) *sanction.Params {
	rv := sanction.DefaultParams()

	k.IterateParams(ctx, func(name, value string) bool {
		switch name {
		case ParamNameImmediateSanctionMinDeposit:
			rv.ImmediateSanctionMinDeposit = toCoinsOrDefault(value, rv.ImmediateSanctionMinDeposit)
		case ParamNameImmediateUnsanctionMinDeposit:
			rv.ImmediateUnsanctionMinDeposit = toCoinsOrDefault(value, rv.ImmediateUnsanctionMinDeposit)
		default:
			panic(fmt.Errorf("unknown param key: %q", name))
		}
		return false
	})

	return rv
}

// SetParams sets the sanction module's params.
// Providing a nil params will cause all params to be deleted (so that defaults are used).
func (k Keeper) SetParams(ctx sdk.Context, params *sanction.Params) {
	store := ctx.KVStore(k.storeKey)
	if params == nil {
		k.deleteParam(store, ParamNameImmediateSanctionMinDeposit)
		k.deleteParam(store, ParamNameImmediateUnsanctionMinDeposit)
	} else {
		k.setParam(store, ParamNameImmediateSanctionMinDeposit, params.ImmediateSanctionMinDeposit.String())
		k.setParam(store, ParamNameImmediateUnsanctionMinDeposit, params.ImmediateUnsanctionMinDeposit.String())
	}
}

// IterateParams iterates over all params entries.
// The callback takes in the name and value, and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateParams(ctx sdk.Context, cb func(name, value string) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), ParamsPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(string(iter.Key()), string(iter.Value())) {
			break
		}
	}
}

// GetImmediateSanctionMinDeposit gets the minimum deposit for a sanction to happen immediately.
func (k Keeper) GetImmediateSanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	return k.getParamAsCoinsOrDefault(
		ctx,
		ParamNameImmediateSanctionMinDeposit,
		sanction.DefaultImmediateSanctionMinDeposit,
	)
}

// GetImmediateUnsanctionMinDeposit gets the minimum deposit for an unsanction to happen immediately.
func (k Keeper) GetImmediateUnsanctionMinDeposit(ctx sdk.Context) sdk.Coins {
	return k.getParamAsCoinsOrDefault(
		ctx,
		ParamNameImmediateUnsanctionMinDeposit,
		sanction.DefaultImmediateUnsanctionMinDeposit,
	)
}

// getParam returns a param value and wether it existed.
func (k Keeper) getParam(store sdk.KVStore, name string) (string, bool) {
	key := CreateParamKey(name)
	if store.Has(key) {
		return string(store.Get(key)), true
	}
	return "", false
}

// setParam sets a param value.
func (k Keeper) setParam(store sdk.KVStore, name, value string) {
	key := CreateParamKey(name)
	val := []byte(value)
	store.Set(key, val)
}

// deleteParam deletes a param value.
func (k Keeper) deleteParam(store sdk.KVStore, name string) {
	key := CreateParamKey(name)
	store.Delete(key)
}

// getParamAsCoinsOrDefault gets a param value and converts it to a coins if possible.
// If the param doesn't exist, the default is returned.
// If the param's value cannot be converted to a Coins, the default is returned.
func (k Keeper) getParamAsCoinsOrDefault(ctx sdk.Context, name string, dflt sdk.Coins) sdk.Coins {
	coins, has := k.getParam(ctx.KVStore(k.storeKey), name)
	if !has {
		return dflt
	}
	return toCoinsOrDefault(coins, dflt)
}

// toCoinsOrDefault converts a string to coins if possible or else returns the provided default.
func toCoinsOrDefault(coins string, dflt sdk.Coins) sdk.Coins {
	rv, err := sdk.ParseCoinsNormalized(coins)
	if err != nil {
		return dflt
	}
	return rv
}
