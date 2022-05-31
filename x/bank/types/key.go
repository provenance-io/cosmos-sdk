package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "bank"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore keys
var (
	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix      = []byte{0x02}
	SupplyKey           = []byte{0x00}
	DenomMetadataPrefix = []byte{0x1}

	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = []byte{0x04}
)

const (
	// TrueB is a byte with value 1 that represents true.
	TrueB = byte(0x01)
	// FalseB is a byte with value 0 that represents false.
	FalseB = byte(0x00)
)

// DenomMetadataKey returns the denomination metadata key.
func DenomMetadataKey(denom string) []byte {
	d := []byte(denom)
	return append(DenomMetadataPrefix, d...)
}

// AddressFromBalancesStore returns an account address from a balances prefix
// store. The key must not contain the prefix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
//
// If invalid key is passed, AddressFromBalancesStore returns ErrInvalidKey.
func AddressFromBalancesStore(key []byte) (sdk.AccAddress, error) {
	if len(key) == 0 {
		return nil, ErrInvalidKey
	}
	addrLen := key[0]
	bound := int(addrLen)
	if len(key)-1 < bound {
		return nil, ErrInvalidKey
	}
	return key[1 : bound+1], nil
}

// CreateAccountBalancesPrefix creates the prefix for an account's balances.
func CreateAccountBalancesPrefix(addr []byte) []byte {
	return append(BalancesPrefix, address.MustLengthPrefix(addr)...)
}

// CreateSendEnabledKey creates the key of the SendDisabled flag for a denom.
func CreateSendEnabledKey(denom string) []byte {
	key := make([]byte, len(SendEnabledPrefix)+len(denom))
	copy(key, SendEnabledPrefix)
	copy(key[len(SendEnabledPrefix):], denom)
	return key
}

// IsTrueB returns true if the provided byte slice has exactly one byte, and it is equal to TrueB.
func IsTrueB(bz []byte) bool {
	return len(bz) == 1 && bz[0] == TrueB
}

// ToBoolB returns TrueB if v is true, and FalseB if it's false.
func ToBoolB(v bool) byte {
	if v {
		return TrueB
	}
	return FalseB
}
