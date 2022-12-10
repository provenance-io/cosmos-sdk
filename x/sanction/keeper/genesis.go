package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// InitGenesis updates this keeper's store using the provided GenesisState.
func (k Keeper) InitGenesis(origCtx sdk.Context, genState *sanction.GenesisState) {
	// We don't want the events from this, so use a context with a throw-away event manager.
	ctx := origCtx.WithEventManager(sdk.NewEventManager())
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	toSanction, err := toAccAddrs(genState.SanctionedAddresses)
	if err != nil {
		panic(err)
	}
	err = k.SanctionAddresses(ctx, toSanction...)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis reads this keeper's entire state and returns it as a GenesisState.
func (k Keeper) ExportGenesis(ctx sdk.Context) *sanction.GenesisState {
	params := k.GetParams(ctx)
	sanctionedAddrs := k.GetAllSanctionedAddresses(ctx)
	return sanction.NewGenesisState(params, sanctionedAddrs)
}

// GetAllSanctionedAddresses gets the bech32 string of every account that is sanctioned.
// This is designed for use with ExportGenesis. See also IterateSanctionedAddresses.
func (k Keeper) GetAllSanctionedAddresses(ctx sdk.Context) []string {
	var rv []string
	k.IterateSanctionedAddresses(ctx, func(addr sdk.AccAddress) bool {
		rv = append(rv, addr.String())
		return false
	})
	return rv
}
