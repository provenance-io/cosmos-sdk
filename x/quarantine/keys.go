package quarantine

import (
	"bytes"
	"crypto/sha256"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// CreateRecordKey creates the key for a quarantine record.
//
// If there is only one fromAddr, it is used as the record suffix.
// If there are more than one, a hash of them is used as the suffix.
// Panics if no fromAddrs are provided.
//
func CreateRecordKey(toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	toAddrPreBz := CreateRecordToAddrPrefix(toAddr)
	recordId := address.MustLengthPrefix(createRecordSuffix(fromAddrs))
	key := make([]byte, len(toAddrPreBz)+len(recordId))
	copy(key, toAddrPreBz)
	copy(key[len(toAddrPreBz):], recordId)
	return key
}

// createRecordSuffix creates a single "address" to use for the provided from addresses.
// This is not to be confused with CreateRecordKey which creates the full key for a quarantine record.
// This only creates a portion of the key.
//
// If one fromAddr is provided, it's what's returned.
// If more than one is provided, they are sorted, combined, and hashed.
//
// Panics if none are provided.
func createRecordSuffix(fromAddrs []sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	switch len(fromAddrs) {
	case 0:
		panic(sdkerrors.ErrLogic.Wrap("at least one fromAddr is required"))
	case 1:
		return fromAddrs[0]
	default:
		// the same n addresses needs to always create the same result.
		sort.Slice(fromAddrs, func(i, j int) bool {
			return bytes.Compare(fromAddrs[i], fromAddrs[j]) < 0
		})
		var toHash []byte
		for _, addr := range fromAddrs {
			toHash = append(toHash, addr...)
		}
		hash := sha256.Sum256(toHash)
		return hash[0:]
	}
}

// ParseRecordKey extracts the to address and record suffix from the provided quarantine funds key.
func ParseRecordKey(key []byte) (toAddr, recordSuffix sdk.AccAddress) {
	// key is of format:
	// 0x22<to addr len><to addr bytes><from addr len><from addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	recordSuffix, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, recordSuffix
}
