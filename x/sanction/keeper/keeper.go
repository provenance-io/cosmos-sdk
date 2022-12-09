package keeper

import (
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

func (k Keeper) DeleteTempEntries(ctx sdk.Context, govPropId uint64, addrs ...sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	for _, addr := range addrs {
		key := CreateTemporaryKey(addr, govPropId)
		store.Delete(key)
	}
}

func (k Keeper) DeleteTempEntriesForAddrs(ctx sdk.Context, addrs ...sdk.AccAddress) {
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

// IterateSanctionedAddresses iterates over all of the permanently sanctioned addresses.
// The callback takes in the sanctioned address and should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateSanctionedAddresses(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), SanctionedPrefix)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr, _ := ParseLengthPrefixedBz(iter.Key())
		if cb(addr) {
			break
		}
	}
}

// IterateTemporaryEntries iterates over each of the temporary entries.
// If an address is provided, only the temporary entries for that address are iterated,
// otherwise all entries are iterated.
// The callback takes in the address in question, the governance proposal associated with it, and whether it's a sanction (true) or unsanction (false).
// The callback should return whether to stop iteration (true = stop, false = keep going).
func (k Keeper) IterateTemporaryEntries(ctx sdk.Context, addr sdk.AccAddress, cb func(addr sdk.AccAddress, govPropId uint64, isSanction bool) (stop bool)) {
	pre := CreateTemporaryAddrPrefix(addr)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), pre)

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
