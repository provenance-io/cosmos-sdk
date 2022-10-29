package quarantine

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the module
	ModuleName = "quarantine"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName
)

var (
	// OptInPrefix is the prefix for the quarantine account opt-in flags.
	OptInPrefix = []byte{0x00}

	// AutoResponsePrefix is the prefix for quarantine auto-response settings.
	AutoResponsePrefix = []byte{0x01}

	// RecordPrefix is the prefix for keys with the records of quarantined funds.
	RecordPrefix = []byte{0x02}
)

// CreateOptInKey creates the key for a quarantine opt-in record.
func CreateOptInKey(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(OptInPrefix)+len(toAddrBz))
	copy(key, OptInPrefix)
	copy(key[len(OptInPrefix):], toAddrBz)
	return key
}

// ParseOptInKey extracts the account address from the provided quarantine opt-in key.
func ParseOptInKey(key []byte) sdk.AccAddress {
	// key is of format:
	// 0x20<to addr len><to addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, _ := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	return toAddr
}

// CreateAutoResponseToAddrPrefix creates a prefix for the quarantine auto-responses for a receiving address.
func CreateAutoResponseToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(AutoResponsePrefix)+len(toAddrBz))
	copy(key, AutoResponsePrefix)
	copy(key[len(AutoResponsePrefix):], toAddrBz)
	return key
}

// CreateAutoResponseKey creates the key for a quarantine auto-response.
func CreateAutoResponseKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateAutoResponseToAddrPrefix(toAddr)
	fromAddrBz := address.MustLengthPrefix(fromAddr)
	key := make([]byte, len(toAddrPreBz)+len(fromAddrBz))
	copy(key, toAddrPreBz)
	copy(key[len(toAddrPreBz):], fromAddrBz)
	return key
}

// ParseAutoResponseKey extracts the to address and from address from the provided quarantine auto-response key.
func ParseAutoResponseKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x21<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

// CreateRecordToAddrPrefix creates a prefix for the quarantine funds for a receiving address.
func CreateRecordToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(RecordPrefix)+len(toAddrBz))
	copy(key, RecordPrefix)
	copy(key[len(RecordPrefix):], toAddrBz)
	return key
}

// CreateRecordKey creates the key for quarantine funds.
func CreateRecordKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateRecordToAddrPrefix(toAddr)
	fromAddrBz := address.MustLengthPrefix(fromAddr)
	key := make([]byte, len(toAddrPreBz)+len(fromAddrBz))
	copy(key, toAddrPreBz)
	copy(key[len(toAddrPreBz):], fromAddrBz)
	return key
}

// ParseRecordKey extracts the to address and from address from the provided quarantine funds key.
func ParseRecordKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x22<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}
