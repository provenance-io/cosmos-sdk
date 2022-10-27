package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// QuarantineKeeper defines a module interface that facilitates management of account and fund quarantines.
type QuarantineKeeper interface {
	GetQuarantinedFundsHolder() sdk.AccAddress

	IsQuarantined(ctx sdk.Context, toAddr sdk.AccAddress) bool
	SetQuarantineOptIn(ctx sdk.Context, toAddr sdk.AccAddress)
	SetQuarantineOptOut(ctx sdk.Context, toAddr sdk.AccAddress)
	IterateQuarantinedAccounts(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool))

	GetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) types.QuarantineAutoResponse
	SetQuarantineAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse)
	IterateQuarantineAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse) (stop bool))

	GetQuarantineRecord(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) types.QuarantineRecord
	SetQuarantineRecord(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, funds *types.QuarantineRecord)
	AddQuarantinedCoins(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, coins sdk.Coins)
	IterateQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, funds types.QuarantineRecord) (stop bool))
	SetQuarantineRecordAccepted(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress)
	SetQuarantineRecordDeclined(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress)
}

var _ QuarantineKeeper = (*BaseQuarantineKeeper)(nil)

// BaseQuarantineKeeper manages quarantine account data.
type BaseQuarantineKeeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	quarantinedFundsHolder sdk.AccAddress
}

func NewBaseQuarantineKeeper(
	cdc codec.BinaryCodec, storeKey storetypes.StoreKey, quarantinedFundsHolder sdk.AccAddress,
) BaseQuarantineKeeper {
	if len(quarantinedFundsHolder) == 0 {
		quarantinedFundsHolder = authtypes.NewModuleAddress(types.ModuleName)
	}
	return BaseQuarantineKeeper{
		cdc:                    cdc,
		storeKey:               storeKey,
		quarantinedFundsHolder: quarantinedFundsHolder,
	}
}

// GetQuarantinedFundsHolder returns the account address that holds quarantined funds.
func (k BaseQuarantineKeeper) GetQuarantinedFundsHolder() sdk.AccAddress {
	return k.quarantinedFundsHolder
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

// getQuarantinedAccountsPrefixStore returns a kv store prefixed for quarantine opt-in entries.
func (k BaseQuarantineKeeper) getQuarantinedAccountsPrefixStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.QuarantineOptInPrefix)
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
// The callback function should accept the to address (that has quarantine enabled).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k BaseQuarantineKeeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(toAddr sdk.AccAddress) (stop bool)) {
	store := k.getQuarantinedAccountsPrefixStore(ctx)
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

// getQuarantineAutoResponsesPrefixStore returns a kv store prefixed for quarantine auto-responses to the given address.
// If toAddr is empty, it will be prefixed for all quarantine auto-responses.
func (k BaseQuarantineKeeper) getQuarantineAutoResponsesPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) sdk.KVStore {
	pre := types.QuarantineAutoResponsePrefix
	if len(toAddr) > 0 {
		pre = types.CreateQuarantineAutoResponseToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateQuarantineAutoResponses iterates over the auto-responses for a given recipient address,
// or if no address is provided, iterates over all auto-response entries.
// The callback function should accept a to address, from address, and auto-response setting (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k BaseQuarantineKeeper) IterateQuarantineAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response types.QuarantineAutoResponse) (stop bool)) {
	store := k.getQuarantineAutoResponsesPrefixStore(ctx, toAddr)
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

// GetQuarantineRecord gets the funds that are quarantined to toAddr from fromAddr.
// If there are no such funds, this will return a QuarantineRecord with zero coins.
func (k BaseQuarantineKeeper) GetQuarantineRecord(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) types.QuarantineRecord {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineRecordKey(toAddr, fromAddr)
	bz := store.Get(key)
	return k.mustBzToQuarantineRecord(bz)
}

// SetQuarantineRecord sets a quarantined funds entry.
// If funds is nil, this will delete any existing entry.
func (k BaseQuarantineKeeper) SetQuarantineRecord(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, funds *types.QuarantineRecord) {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateQuarantineRecordKey(toAddr, fromAddr)
	if funds == nil {
		store.Delete(key)
	} else {
		val := k.cdc.MustMarshal(funds)
		store.Set(key, val)
	}
}

// AddQuarantinedCoins records that some new funds have been quarantined.
func (k BaseQuarantineKeeper) AddQuarantinedCoins(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, coins sdk.Coins) {
	qf := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
	qf.Add(coins...)
	qf.Declined = k.GetQuarantineAutoResponse(ctx, toAddr, fromAddr) == types.QUARANTINE_AUTO_RESPONSE_DECLINE
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, &qf)
}

// getQuarantineRecordPrefixStore returns a kv store prefixed for quarantine records to the given address.
// If toAddr is empty, it will be prefixed for all quarantine records.
func (k BaseQuarantineKeeper) getQuarantineRecordPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) sdk.KVStore {
	pre := types.QuarantineRecordPrefix
	if len(toAddr) > 0 {
		pre = types.CreateQuarantineRecordToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateQuarantineRecords iterates over the quarantined funds for a given recipient address,
// or if no address is provided, iterates over all quarantined funds.
// The callback function should accept a to address, from address, and QuarantineRecord (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k BaseQuarantineKeeper) IterateQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, funds types.QuarantineRecord) (stop bool)) {
	store := k.getQuarantineRecordPrefixStore(ctx, toAddr)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kFromAddr := types.ParseQuarantineRecordKey(iter.Key())
		qf := k.mustBzToQuarantineRecord(iter.Value())

		if cb(kToAddr, kFromAddr, qf) {
			break
		}
	}
}

// SetQuarantineRecordAccepted marks quarantined funds as accepted.
func (k BaseQuarantineKeeper) SetQuarantineRecordAccepted(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, nil)
}

// SetQuarantineRecordDeclined marks some quarantined funds as declined.
func (k BaseQuarantineKeeper) SetQuarantineRecordDeclined(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	qf := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
	qf.Declined = true
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, &qf)
}

// bzToQuarantineRecord converts the given byte slice into a QuarantineRecord or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecord is returned with zero coins.
func (k BaseQuarantineKeeper) bzToQuarantineRecord(bz []byte) (types.QuarantineRecord, error) {
	qf := types.QuarantineRecord{
		Coins:    sdk.Coins{},
		Declined: false,
	}
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &qf)
		if err != nil {
			return qf, err
		}
	}
	return qf, nil
}

// mustBzToQuarantineRecord returns bzToQuarantineRecord but panics on error.
func (k BaseQuarantineKeeper) mustBzToQuarantineRecord(bz []byte) types.QuarantineRecord {
	qf, err := k.bzToQuarantineRecord(bz)
	if err != nil {
		panic(err)
	}
	return qf
}
