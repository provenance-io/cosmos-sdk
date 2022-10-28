package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		// TODO[1046]: Finish NewDecodeStore. A case for each prefix.
		default:
			panic(fmt.Sprintf("invalid quarantine key %X", kvA.Key))
		}
	}
}
