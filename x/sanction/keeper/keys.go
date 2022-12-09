package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// Keys for store prefixes
// Items are stored with the following keys:
//
// Sanctioned addresses:
// - 0x00<addr len (1 byte)><addr> -> 0x00
// Temporarily sanctioned addresses:
// - 0x01<addr len (1 byte)><addr><gov prop id> -> 0x01
// Temporarily unsanctioned addresses:
// - 0x01<addr len (1 byte)><addr><gov prop id> -> 0x00
// Params entry:
// - 0x02<name> -> <value>
var (
	SanctionedPrefix = []byte{0x00}
	TemporaryPrefix  = []byte{0x01}
	ParamsPrefix     = []byte{0x02}
)

const (
	ParamNameImmediateSanctionMinDeposit   = "immediate_sanction_min_deposit"
	ParamNameImmediateUnsanctionMinDeposit = "immediate_unsanction_min_deposit"
)

// ConcatBz creates a single byte slice consisting of the two provided byte slices.
func ConcatBz(bz1, bz2 []byte) []byte {
	rv := make([]byte, len(bz1)+len(bz2))
	copy(rv, bz1)
	copy(rv[len(bz1):], bz2)
	return rv
}

// ParseLengthPrefixedBz parses a length-prefixed byte slice into those bytes and any leftover bytes.
func ParseLengthPrefixedBz(bz []byte) ([]byte, []byte) {
	addrLen, addrLenEndIndex := sdk.ParseLengthPrefixedBytes(bz, 0, 1)
	addr, addrEndIndex := sdk.ParseLengthPrefixedBytes(bz, addrLenEndIndex+1, int(addrLen[0]))
	var remainder []byte
	if len(bz) > addrEndIndex+1 {
		remainder = bz[addrEndIndex+1:]
	}
	return addr, remainder
}

// CreateSanctionedAddrKey creates the sanctioned address key for the provided address.
//
// - 0x00<addr len (1 byte)><addr>
func CreateSanctionedAddrKey(addr sdk.AccAddress) []byte {
	return ConcatBz(SanctionedPrefix, address.MustLengthPrefix(addr))
}

// ParseSanctionedAddrKey extracts the address from the provided sanctioned address key.
func ParseSanctionedAddrKey(key []byte) sdk.AccAddress {
	addr, _ := ParseLengthPrefixedBz(key[1:])
	return addr
}

// CreateTemporaryAddrPrefix creates a key prefix for a temporarily sanctioned/unsanctioned address.
//
// - 0x01<addr len(1 byte)><addr>
func CreateTemporaryAddrPrefix(addr sdk.AccAddress) []byte {
	return ConcatBz(TemporaryPrefix, address.MustLengthPrefix(addr))
}

// CreateTemporaryKey creates a key for a temporarily sanctioned/unsanctioned address associated with the given governance proposal id.
//
// - 0x01<addr len (1 byte)><addr><gov prop id>
func CreateTemporaryKey(addr sdk.AccAddress, govPropId uint64) []byte {
	return ConcatBz(CreateTemporaryAddrPrefix(addr), sdk.Uint64ToBigEndian(govPropId))
}

// ParseTemporaryKey extracts the address and gov prop id from the provided temporary key.
func ParseTemporaryKey(key []byte) (sdk.AccAddress, uint64) {
	addr, govPropIdBz := ParseLengthPrefixedBz(key[1:])
	govPropId := sdk.BigEndianToUint64(govPropIdBz)
	return addr, govPropId
}

const (
	// TempSanctionB is a byte representing a temporary sanction.
	TempSanctionB = 0x01
	// TempUnsanctionB is a byte representing a temporary unsanction.
	TempUnsanctionB = 0x00
)

// IsTempSanctionBz returns true if the provided byte slice indicates a temporary sanction.
func IsTempSanctionBz(bz []byte) bool {
	return len(bz) == 1 && bz[0] == TempSanctionB
}

// IsTempUnsanctionBz returns true if the provided byte slice indicates a temporary unsanction.
func IsTempUnsanctionBz(bz []byte) bool {
	return len(bz) == 1 && bz[0] == TempUnsanctionB
}

func CreateParamKey(name string) []byte {
	return ConcatBz(ParamsPrefix, []byte(name))
}

func ParseParamKey(bz []byte) string {
	return string(bz[1:])
}
