package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
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
	SupplyKey           = []byte{0x00}
	DenomMetadataPrefix = []byte{0x1}
	DenomAddressPrefix  = []byte{0x03}

	// BalancesPrefix is the prefix for the account balances store. We use a byte
	// (instead of `[]byte("balances")` to save some disk space).
	BalancesPrefix = []byte{0x02}

	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = []byte{0x04}

	// QuarantineOptInPrefix is the prefix for the quarantine account opt-in flags.
	QuarantineOptInPrefix = []byte{0x20}

	// QuarantineAutoResponsePrefix is the prefix for quarantine auto-response settings.
	QuarantineAutoResponsePrefix = []byte{0x21}
)

const (
	// TrueB is a byte with value 1 that represents true.
	TrueB = byte(0x01)
	// FalseB is a byte with value 0 that represents false.
	FalseB = byte(0x00)

	// NoAutoB is a byte with value 0 (corresponding to QUARANTINE_AUTO_RESPONSE_UNSPECIFIED).
	NoAutoB = byte(0x00)
	// AutoAcceptB is a byte with value 1 (corresponding to QUARANTINE_AUTO_RESPONSE_ACCEPT).
	AutoAcceptB = byte(0x01)
	// AutoDeclineB is a byte with value 2 (corresponding to QUARANTINE_AUTO_RESPONSE_DECLINE).
	AutoDeclineB = byte(0x02)
)

// AddressAndDenomFromBalancesStore returns an account address and denom from a balances prefix
// store. The key must not contain the prefix BalancesPrefix as the prefix store
// iterator discards the actual prefix.
//
// If invalid key is passed, AddressAndDenomFromBalancesStore returns ErrInvalidKey.
func AddressAndDenomFromBalancesStore(key []byte) (sdk.AccAddress, string, error) {
	if len(key) == 0 {
		return nil, "", ErrInvalidKey
	}

	kv.AssertKeyAtLeastLength(key, 1)

	addrBound := int(key[0])

	if len(key)-1 < addrBound {
		return nil, "", ErrInvalidKey
	}

	return key[1 : addrBound+1], string(key[addrBound+1:]), nil
}

// CreatePrefixedAccountStoreKey returns the key for the given account and denomination.
// This method can be used when performing an ABCI query for the balance of an account.
func CreatePrefixedAccountStoreKey(addr []byte, denom []byte) []byte {
	return append(CreateAccountBalancesPrefix(addr), denom...)
}

// CreateAccountBalancesPrefix creates the prefix for an account's balances.
func CreateAccountBalancesPrefix(addr []byte) []byte {
	return append(BalancesPrefix, address.MustLengthPrefix(addr)...)
}

// CreateDenomAddressPrefix creates a prefix for a reverse index of denomination
// to account balance for that denomination.
func CreateDenomAddressPrefix(denom string) []byte {
	// we add a "zero" byte at the end - null byte terminator, to allow prefix denom prefix
	// scan. Setting it is not needed (key[last] = 0) - because this is the default.
	key := make([]byte, len(DenomAddressPrefix)+len(denom)+1)
	copy(key, DenomAddressPrefix)
	copy(key[len(DenomAddressPrefix):], denom)
	return key
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

// CreateQuarantineOptInKey creates the key for a quarantine opt-in record.
func CreateQuarantineOptInKey(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(QuarantineOptInPrefix)+len(toAddrBz))
	copy(key, QuarantineOptInPrefix)
	copy(key[len(QuarantineOptInPrefix):], toAddrBz)
	return key
}

// ParseQuarantineOptInKey extracts the account address from the provided quarantine opt-in key.
func ParseQuarantineOptInKey(key []byte) sdk.AccAddress {
	// key is of format:
	// 0x20<to addr len><to addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, _ := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	return toAddr
}

// CreateQuarantineAutoResponseToAddrPrefix creates a prefix for the quarantine auto-responses for a receiving address.
func CreateQuarantineAutoResponseToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(QuarantineAutoResponsePrefix)+len(toAddrBz))
	copy(key, QuarantineAutoResponsePrefix)
	copy(key[len(QuarantineAutoResponsePrefix):], toAddrBz)
	return key
}

// CreateQuarantineAutoResponseKey creates the key for a quarantine auto-response.
func CreateQuarantineAutoResponseKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateQuarantineAutoResponseToAddrPrefix(toAddr)
	fromAddrBz := address.MustLengthPrefix(fromAddr)
	key := make([]byte, len(toAddrPreBz)+len(fromAddrBz))
	copy(key, toAddrPreBz)
	copy(key[len(toAddrPreBz):], fromAddrBz)
	return key
}

// ParseQuarantineAutoResponseKey extracts the to address and from address from the provided quarantine auto-response key.
func ParseQuarantineAutoResponseKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x20<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

// ToAutoB converts a QuarantineAutoResponse into the byte that will represent it.
func ToAutoB(r QuarantineAutoResponse) byte {
	switch r {
	case QUARANTINE_AUTO_RESPONSE_ACCEPT:
		return AutoAcceptB
	case QUARANTINE_AUTO_RESPONSE_DECLINE:
		return AutoDeclineB
	default:
		return NoAutoB
	}
}

// ToQuarantineAutoResponse returns the QuarantineAutoResponse represented by the provided byte slice.
func ToQuarantineAutoResponse(bz []byte) QuarantineAutoResponse {
	if len(bz) == 1 {
		switch bz[0] {
		case AutoAcceptB:
			return QUARANTINE_AUTO_RESPONSE_ACCEPT
		case AutoDeclineB:
			return QUARANTINE_AUTO_RESPONSE_DECLINE
		}
	}
	return QUARANTINE_AUTO_RESPONSE_UNSPECIFIED
}

// IsAutoAcceptB returns true if the provided byte slice has exactly one byte, and it is equal to AutoAccept.
func IsAutoAcceptB(bz []byte) bool {
	return len(bz) == 1 && bz[0] == AutoAcceptB
}

// IsAutoDeclineB returns true if the provided byte slice has exactly one byte, and it is equal to AutoDecline.
func IsAutoDeclineB(bz []byte) bool {
	return len(bz) == 1 && bz[0] == AutoDeclineB
}
