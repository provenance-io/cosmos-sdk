package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, quarantine.QuarantineOptInPrefix):
			// The values are all supposed to be [0x00]. So just output the raw byte slice.
			return fmt.Sprintf("AddrA: %v\nAddrB: %v", kvA.Value, kvA.Value)

		case bytes.HasPrefix(kvA.Key, quarantine.QuarantineAutoResponsePrefix):
			respA := quarantine.ToQuarantineAutoResponse(kvA.Value)
			respB := quarantine.ToQuarantineAutoResponse(kvB.Value)
			return fmt.Sprintf("%s\n%s", respA.String(), respB.String())

		case bytes.HasPrefix(kvA.Key, quarantine.QuarantineRecordPrefix):
			var qrA, qrB quarantine.QuarantineRecord
			cdc.MustUnmarshal(kvA.Value, &qrA)
			cdc.MustUnmarshal(kvB.Value, &qrB)
			return fmt.Sprintf("%v\n%v", qrA, qrB)

		default:
			panic(fmt.Sprintf("invalid quarantine key %X", kvA.Key))
		}
	}
}
