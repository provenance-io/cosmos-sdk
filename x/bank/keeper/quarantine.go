package keeper

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// QuarantineKeeper defines a module interface that facilitates management of account and fund quarantines.
type QuarantineKeeper interface {
	IsQuarantined(ctx sdk.Context, addr sdk.AccAddress) bool

	QuarantineOptIn(ctx sdk.Context, addr sdk.AccAddress)
	QuarantineOptOut(ctx sdk.Context, addr sdk.AccAddress)

	IterateQuarantinedAccounts(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool))
}

var _ QuarantineKeeper = (*BaseQuarantineKeeper)(nil)

// BaseQuarantineKeeper manages quarantine account data.
type BaseQuarantineKeeper struct {
	storeKey storetypes.StoreKey
}

func NewBaseQuarantineKeeper(storeKey storetypes.StoreKey) BaseQuarantineKeeper {
	return BaseQuarantineKeeper{
		storeKey: storeKey,
	}
}

// IsQuarantined returns true if the given address has opted into quarantine.
func (k BaseQuarantineKeeper) IsQuarantined(ctx sdk.Context, addr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(addr)
	return store.Has(key)
}

// QuarantineOptIn records that an address has opted into quarantine.
func (k BaseQuarantineKeeper) QuarantineOptIn(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(addr)
	store.Set(key, []byte{0x00})
}

// QuarantineOptOut removes an address' quarantine opt-in record.
func (k BaseQuarantineKeeper) QuarantineOptOut(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(addr)
	store.Delete(key)
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
func (k BaseQuarantineKeeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr := types.ParseQuarantineOptInKey(iter.Key())
		if cb(addr) {
			break
		}
	}
}
