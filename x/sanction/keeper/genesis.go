package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// InitGenesis updates this keeper's store using the provided GenesisState.
func (k Keeper) InitGenesis(ctx sdk.Context, genesisState *sanction.GenesisState) {
	// TODO[1046]: Implement InitGenesis
	panic("not implemented")
}

// ExportGenesis reads this keeper's entire state and returns it as a GenesisState.
func (k Keeper) ExportGenesis(ctx sdk.Context) *sanction.GenesisState {
	// TODO[1046]: Implement ExportGenesis
	panic("not implemented")
}
