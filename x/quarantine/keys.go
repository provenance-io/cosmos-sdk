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

// TODO[1046]: copy the keys stuff I need into here.

var (
	// QuarantineOptInPrefix is the prefix for the quarantine account opt-in flags.
	QuarantineOptInPrefix = []byte{0x00}

	// QuarantineAutoResponsePrefix is the prefix for quarantine auto-response settings.
	QuarantineAutoResponsePrefix = []byte{0x01}

	// QuarantineRecordPrefix is the prefix for keys with the records of quarantined funds.
	QuarantineRecordPrefix = []byte{0x02}
)

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
	// 0x21<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

// CreateQuarantineRecordToAddrPrefix creates a prefix for the quarantine funds for a receiving address.
func CreateQuarantineRecordToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	key := make([]byte, len(QuarantineRecordPrefix)+len(toAddrBz))
	copy(key, QuarantineRecordPrefix)
	copy(key[len(QuarantineRecordPrefix):], toAddrBz)
	return key
}

// CreateQuarantineRecordKey creates the key for quarantine funds.
func CreateQuarantineRecordKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateQuarantineRecordToAddrPrefix(toAddr)
	fromAddrBz := address.MustLengthPrefix(fromAddr)
	key := make([]byte, len(toAddrPreBz)+len(fromAddrBz))
	copy(key, toAddrPreBz)
	copy(key[len(toAddrPreBz):], fromAddrBz)
	return key
}

// ParseQuarantineRecordKey extracts the to address and from address from the provided quarantine funds key.
func ParseQuarantineRecordKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x22<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}
