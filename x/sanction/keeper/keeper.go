package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper sanction.BankKeeper

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper sanction.BankKeeper, authority string) Keeper {
	rv := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
		authority:  authority,
	}
	bankKeeper.SetSanctionKeeper(rv)
	return rv
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) IsSanctionedAddr(ctx sdk.Context, addr sdk.AccAddress) bool {
	// TODO[1046]: Implement IsSanctioned
	panic("not implemented")
}

// TODO[1046]: Implement the rest of the needed keeper functions.
