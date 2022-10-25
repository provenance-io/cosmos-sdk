package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// QuarantineKeeper defines a module interface that facilitates management of account and fund quarantines.
type QuarantineKeeper interface {
	IsQuarantined(ctx sdk.Context, toAddr sdk.AccAddress) bool

	SetQuarantineOptIn(ctx sdk.Context, toAddr sdk.AccAddress)
	SetQuarantineOptOut(ctx sdk.Context, toAddr sdk.AccAddress)

	IterateQuarantinedAccounts(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool))

	GetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) types.QuarantineAutoResponse
	SetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse)

	IterateQuarantinedAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse) (stop bool))

	SetQuarantinedFundsAccepted(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress)
	SetQuarantinedFundsDecline(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress)
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
func (k BaseQuarantineKeeper) IsQuarantined(ctx sdk.Context, toAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(toAddr)
	return store.Has(key)
}

// SetQuarantineOptIn records that an address has opted into quarantine.
func (k BaseQuarantineKeeper) SetQuarantineOptIn(ctx sdk.Context, toAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(toAddr)
	store.Set(key, []byte{0x00})
}

// SetQuarantineOptOut removes an address' quarantine opt-in record.
func (k BaseQuarantineKeeper) SetQuarantineOptOut(ctx sdk.Context, toAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineOptInKey(toAddr)
	store.Delete(key)
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
// The callback function should accept the to address (that has quarantine enabled).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k BaseQuarantineKeeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(toAddr sdk.AccAddress) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.QuarantineOptInPrefix)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr := types.ParseQuarantineOptInKey(iter.Key())
		if cb(addr) {
			break
		}
	}
}

// GetQuarantineAutoResponse returns the quarantine auto-response for the given to/from addresses.
func (k BaseQuarantineKeeper) GetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) types.QuarantineAutoResponse {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineAutoResponseKey(toAddr, fromAddr)
	bz := store.Get(key)
	return types.ToQuarantineAutoResponse(bz)
}

// SetQuarantineAutoResponse sets the auto response of sends to toAddr from fromAddr.
// If the response is QUARANTINE_AUTO_RESPONSE_UNSPECIFIED, the auto-response record is deleted,
// otherwise it is created/updated with the given setting.
func (k BaseQuarantineKeeper) SetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineAutoResponseKey(toAddr, fromAddr)
	val := types.ToAutoB(response)
	if val == types.NoAutoB {
		store.Delete(key)
	} else {
		store.Set(key, []byte{val})
	}
}

// IterateQuarantinedAutoResponses iterates over all auto-responses for a given to address, or if no address is provided,
// iterates over all auto-response entries.
// The callback function should accept the to address, from address, and auto-response setting (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k BaseQuarantineKeeper) IterateQuarantinedAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse) (stop bool)) {
	var pre []byte
	if len(toAddr) == 0 {
		pre = types.QuarantineAutoResponsePrefix
	} else {
		pre = types.CreateQuarantineAutoResponseToAddrPrefix(toAddr)
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), pre)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kFromAddr := types.ParseQuarantineAutoResponseKey(iter.Key())
		val := types.ToQuarantineAutoResponse(iter.Value())
		if cb(kToAddr, kFromAddr, val) {
			break
		}
	}
}

func (k BaseQuarantineKeeper) SetQuarantinedFundsAccepted(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	panic("not implemented yet")
}

func (k BaseQuarantineKeeper) SetQuarantinedFundsDecline(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	panic("not implemented yet")
}
