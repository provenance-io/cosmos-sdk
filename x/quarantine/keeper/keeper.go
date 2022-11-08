package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper quarantine.BankKeeper

	fundsHolder sdk.AccAddress
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper quarantine.BankKeeper, fundsHolder sdk.AccAddress) Keeper {
	if len(fundsHolder) == 0 {
		fundsHolder = authtypes.NewModuleAddress(quarantine.ModuleName)
	}
	rv := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		bankKeeper:  bankKeeper,
		fundsHolder: fundsHolder,
	}
	bankKeeper.SetQuarantineKeeper(rv)
	return rv
}

// GetFundsHolder returns the account address that holds quarantined funds.
func (k Keeper) GetFundsHolder() sdk.AccAddress {
	return k.fundsHolder
}

// IsQuarantinedAddr returns true if the given address has opted into quarantine.
func (k Keeper) IsQuarantinedAddr(ctx sdk.Context, toAddr sdk.AccAddress) bool {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// SetOptIn records that an address has opted into quarantine.
func (k Keeper) SetOptIn(ctx sdk.Context, toAddr sdk.AccAddress) error {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	store.Set(key, []byte{0x00})
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptIn{ToAddress: toAddr.String()})
}

// SetOptOut removes an address' quarantine opt-in record.
func (k Keeper) SetOptOut(ctx sdk.Context, toAddr sdk.AccAddress) error {
	key := quarantine.CreateOptInKey(toAddr)
	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventOptOut{ToAddress: toAddr.String()})
}

// getQuarantinedAccountsPrefixStore returns a kv store prefixed for quarantine opt-in entries.
func (k Keeper) getQuarantinedAccountsPrefixStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), quarantine.OptInPrefix)
}

// IterateQuarantinedAccounts iterates over all quarantine account addresses.
// The callback function should accept the to address (that has quarantine enabled).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantinedAccounts(ctx sdk.Context, cb func(toAddr sdk.AccAddress) (stop bool)) {
	store := k.getQuarantinedAccountsPrefixStore(ctx)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		addr := quarantine.ParseOptInKey(iter.Key())
		if cb(addr) {
			break
		}
	}
}

// GetAllQuarantinedAccounts gets the bech32 string of every account that have opted into quarantine.
func (k Keeper) GetAllQuarantinedAccounts(ctx sdk.Context) []string {
	var rv []string
	k.IterateQuarantinedAccounts(ctx, func(toAddr sdk.AccAddress) bool {
		rv = append(rv, toAddr.String())
		return false
	})
	return rv
}

// GetAutoResponse returns the quarantine auto-response for the given to/from addresses.
func (k Keeper) GetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) quarantine.AutoResponse {
	key := quarantine.CreateAutoResponseKey(toAddr, fromAddr)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	return quarantine.ToAutoResponse(bz)
}

// IsAutoAccept returns true if the to address has enabled auto-accept from the from address.
func (k Keeper) IsAutoAccept(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) bool {
	return k.GetAutoResponse(ctx, toAddr, fromAddr).IsAccept()
}

// IsAutoDecline returns true if the to address has enabled auto-decline from the from address.
func (k Keeper) IsAutoDecline(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) bool {
	return k.GetAutoResponse(ctx, toAddr, fromAddr).IsDecline()
}

// SetAutoResponse sets the auto response of sends to toAddr from fromAddr.
// If the response is AUTO_RESPONSE_UNSPECIFIED, the auto-response record is deleted,
// otherwise it is created/updated with the given setting.
func (k Keeper) SetAutoResponse(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) {
	key := quarantine.CreateAutoResponseKey(toAddr, fromAddr)
	val := quarantine.ToAutoB(response)
	store := ctx.KVStore(k.storeKey)
	if val == quarantine.NoAutoB {
		store.Delete(key)
	} else {
		store.Set(key, []byte{val})
	}
}

// getAutoResponsesPrefixStore returns a kv store prefixed for quarantine auto-responses to the given address.
// If toAddr is empty, it will be prefixed for all quarantine auto-responses.
func (k Keeper) getAutoResponsesPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) sdk.KVStore {
	pre := quarantine.AutoResponsePrefix
	if len(toAddr) > 0 {
		pre = quarantine.CreateAutoResponseToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateAutoResponses iterates over the auto-responses for a given recipient address,
// or if no address is provided, iterates over all auto-response entries.
// The callback function should accept a to address, from address, and auto-response setting (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateAutoResponses(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) (stop bool)) {
	store := k.getAutoResponsesPrefixStore(ctx, toAddr)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kFromAddr := quarantine.ParseAutoResponseKey(iter.Key())
		val := quarantine.ToAutoResponse(iter.Value())
		if cb(kToAddr, kFromAddr, val) {
			break
		}
	}
}

// GetAllAutoResponseEntries gets a AutoResponseEntry entry for every quarantine auto-response that has been set.
func (k Keeper) GetAllAutoResponseEntries(ctx sdk.Context) []*quarantine.AutoResponseEntry {
	var rv []*quarantine.AutoResponseEntry
	k.IterateAutoResponses(ctx, nil, func(toAddr, fromAddr sdk.AccAddress, resp quarantine.AutoResponse) bool {
		rv = append(rv, quarantine.NewAutoResponseEntry(toAddr, fromAddr, resp))
		return false
	})
	return rv
}

// GetQuarantineRecords gets all the quarantine records to toAddr that involved any of the fromAddrs.
func (k Keeper) GetQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []*quarantine.QuarantineRecord {
	var rv []*quarantine.QuarantineRecord
	allSuffixes := k.getQuarantineRecordSuffixes(ctx, toAddr, fromAddrs...)
	store := ctx.KVStore(k.storeKey)
	for _, suffix := range allSuffixes {
		key := quarantine.CreateRecordKey(toAddr, suffix)
		if store.Has(key) {
			bz := store.Get(key)
			rv = append(rv, k.mustBzToQuarantineRecord(bz))
		}
	}
	return rv
}

// SetQuarantineRecord sets a quarantined funds entry.
// If funds is nil, this will delete any existing entry.
func (k Keeper) SetQuarantineRecord(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, funds *quarantine.QuarantineRecord) {
	key := quarantine.CreateRecordKey(toAddr, fromAddr)
	store := ctx.KVStore(k.storeKey)
	if funds == nil {
		store.Delete(key)
	} else {
		val := k.cdc.MustMarshal(funds)
		store.Set(key, val)
	}
}

// AddQuarantinedCoins records that some new funds have been quarantined.
func (k Keeper) AddQuarantinedCoins(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, coins sdk.Coins) error {
	// TODO[1046]: Refactor this to account for multiple from addresses.
	qf := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
	qf.AddCoins(coins...)
	qf.Declined = k.IsAutoDecline(ctx, toAddr, fromAddr)
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, &qf)
	return ctx.EventManager().EmitTypedEvent(&quarantine.EventFundsQuarantined{
		ToAddress: toAddr.String(),
		Coins:     coins,
	})
}

// getQuarantineRecordPrefixStore returns a kv store prefixed for quarantine records to the given address.
// If toAddr is empty, it will be prefixed for all quarantine records.
func (k Keeper) getQuarantineRecordPrefixStore(ctx sdk.Context, toAddr sdk.AccAddress) sdk.KVStore {
	pre := quarantine.RecordPrefix
	if len(toAddr) > 0 {
		pre = quarantine.CreateRecordToAddrPrefix(toAddr)
	}
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateQuarantineRecords iterates over the quarantined funds for a given recipient address,
// or if no address is provided, iterates over all quarantined funds.
// The callback function should accept a to address, record suffix, and QuarantineRecord (in that order).
// It should return whether to stop iteration early. I.e. false will allow iteration to continue, true will stop iteration.
func (k Keeper) IterateQuarantineRecords(ctx sdk.Context, toAddr sdk.AccAddress, cb func(toAddr, recordSuffix sdk.AccAddress, funds *quarantine.QuarantineRecord) (stop bool)) {
	store := k.getQuarantineRecordPrefixStore(ctx, toAddr)
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kToAddr, kRecordSuffix := quarantine.ParseRecordKey(iter.Key())
		qf := k.mustBzToQuarantineRecord(iter.Value())

		if cb(kToAddr, kRecordSuffix, qf) {
			break
		}
	}
}

// GetAllQuarantinedFunds gets a QuarantinedFunds entry for each QuarantineRecord.
func (k Keeper) GetAllQuarantinedFunds(ctx sdk.Context) []*quarantine.QuarantinedFunds {
	var rv []*quarantine.QuarantinedFunds
	k.IterateQuarantineRecords(ctx, nil, func(toAddr, _ sdk.AccAddress, funds *quarantine.QuarantineRecord) bool {
		rv = append(rv, funds.AsQuarantinedFunds(toAddr))
		return false
	})
	return rv
}

// SetQuarantineRecordAccepted marks quarantined funds as accepted.
func (k Keeper) SetQuarantineRecordAccepted(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, nil)
}

// SetQuarantineRecordDeclined marks some quarantined funds as declined.
func (k Keeper) SetQuarantineRecordDeclined(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress) {
	// TODO[1046]: Refactor this to account for multiple from addresses.
	qf := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
	qf.Declined = true
	k.SetQuarantineRecord(ctx, toAddr, fromAddr, &qf)
}

// bzToQuarantineRecord converts the given byte slice into a QuarantineRecord or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecord is returned with zero coins.
func (k Keeper) bzToQuarantineRecord(bz []byte) (*quarantine.QuarantineRecord, error) {
	qf := quarantine.QuarantineRecord{
		Coins: sdk.Coins{},
	}
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &qf)
		if err != nil {
			return &qf, err
		}
	}
	return &qf, nil
}

// mustBzToQuarantineRecord returns bzToQuarantineRecord but panics on error.
func (k Keeper) mustBzToQuarantineRecord(bz []byte) *quarantine.QuarantineRecord {
	qf, err := k.bzToQuarantineRecord(bz)
	if err != nil {
		panic(err)
	}
	return qf
}

func (k Keeper) setQuarantineRecordSuffixIndex(ctx sdk.Context, toAddr, fromAddr sdk.AccAddress, value *quarantine.QuarantineRecordSuffixIndex) {
	value.RemoveSuffixes(fromAddr)
	value.Simplify()
	key := quarantine.CreateRecordIndexKey(toAddr, fromAddr)
	store := ctx.KVStore(k.storeKey)
	if len(value.RecordSuffixes) == 0 {
		store.Delete(key)
	} else {
		val := k.cdc.MustMarshal(value)
		store.Set(key, val)
	}
}

func (k Keeper) getQuarantineRecordSuffixes(ctx sdk.Context, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) [][]byte {
	store := ctx.KVStore(k.storeKey)
	rv := &quarantine.QuarantineRecordSuffixIndex{}
	for _, fromAddr := range fromAddrs {
		key := quarantine.CreateRecordIndexKey(toAddr, fromAddr)
		bz := store.Get(key)
		suffixes := k.mustBzToQuarantineRecordSuffixIndex(bz)
		rv.AddSuffixes(suffixes.RecordSuffixes...)
		rv.AddSuffixes(fromAddr)
	}
	rv.Simplify()
	return rv.RecordSuffixes
}

// bzToQuarantineRecordSuffixIndex converts the given byte slice into a QuarantineRecordSuffixIndex or returns an error.
// If the byte slice is nil or empty, a default QuarantineRecordSuffixIndex is returned with no suffixes.
func (k Keeper) bzToQuarantineRecordSuffixIndex(bz []byte) (*quarantine.QuarantineRecordSuffixIndex, error) {
	var si quarantine.QuarantineRecordSuffixIndex
	if len(bz) > 0 {
		err := k.cdc.Unmarshal(bz, &si)
		if err != nil {
			return &si, err
		}
	}
	return &si, nil
}

// mustBzToQuarantineRecordSuffixIndex returns bzToQuarantineRecordSuffixIndex but panics on error.
func (k Keeper) mustBzToQuarantineRecordSuffixIndex(bz []byte) *quarantine.QuarantineRecordSuffixIndex {
	si, err := k.bzToQuarantineRecordSuffixIndex(bz)
	if err != nil {
		panic(err)
	}
	return si
}
