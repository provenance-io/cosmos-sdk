package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// This file is available only to unit tests and houses functions for doing
// things with private keeper package stuff.

// WithFundsHolder creates a copy of this setting the funds holder to the provided addr.
func (k Keeper) WithFundsHolder(addr sdk.AccAddress) Keeper {
	k.fundsHolder = addr
	return k
}

// GetCodec gets this keeper's codec.
func (k Keeper) GetCodec() codec.BinaryCodec {
	return k.cdc
}

// BzToQuarantineRecord exposes bzToQuarantineRecord for unit tests.
func (k Keeper) BzToQuarantineRecord(bz []byte) (*quarantine.QuarantineRecord, error) {
	return k.bzToQuarantineRecord(bz)
}

// MustBzToQuarantineRecord exposes mustBzToQuarantineRecord for unit tests.
func (k Keeper) MustBzToQuarantineRecord(bz []byte) *quarantine.QuarantineRecord {
	return k.mustBzToQuarantineRecord(bz)
}
