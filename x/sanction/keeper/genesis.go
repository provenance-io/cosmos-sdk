package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// InitGenesis updates this keeper's store using the provided GenesisState.
func (k Keeper) InitGenesis(ctx sdk.Context, genState *sanction.GenesisState) {
	k.SetParams(ctx, genState.Params)

	var toSanction []sdk.AccAddress
	for _, addr := range genState.SanctionedAddresses {
		toSanction = append(toSanction, sdk.MustAccAddressFromBech32(addr))
	}
	k.SanctionAddresses(ctx, toSanction...)
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
