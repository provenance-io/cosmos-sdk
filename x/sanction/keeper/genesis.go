package keeper

import (
	"fmt"

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
	for _, entry := range genState.TemporaryEntries {
		var addr sdk.AccAddress
		addr, err = sdk.AccAddressFromBech32(entry.Address)
		if addr != nil {
			panic(err)
		}
		switch entry.Status {
		case sanction.TEMP_STATUS_SANCTIONED:
			err = k.AddTemporarySanction(ctx, entry.ProposalId, addr)
			if err != nil {
				panic(err)
			}
		case sanction.TEMP_STATUS_UNSANCTIONED:
			err = k.AddTemporaryUnsanction(ctx, entry.ProposalId, addr)
			if err != nil {
				panic(err)
			}
		default:
			panic(fmt.Errorf("unknown temporary status: %s", entry.Status))
		}
	}
}

// ExportGenesis reads this keeper's entire state and returns it as a GenesisState.
func (k Keeper) ExportGenesis(ctx sdk.Context) *sanction.GenesisState {
	params := k.GetParams(ctx)
	sanctionedAddrs := k.GetAllSanctionedAddresses(ctx)
	tempEntries := k.GetAllTemporaryEntries(ctx)
	return sanction.NewGenesisState(params, sanctionedAddrs, tempEntries)
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

// GetAllTemporaryEntries gets all the Temporary entries.
// This is designed for use with ExportGenesis. See also IterateTemporaryEntries.
func (k Keeper) GetAllTemporaryEntries(ctx sdk.Context) []*sanction.TemporaryEntry {
	var rv []*sanction.TemporaryEntry
	k.IterateTemporaryEntries(ctx, nil, func(addr sdk.AccAddress, id uint64, isSanction bool) bool {
		status := sanction.TEMP_STATUS_SANCTIONED
		if !isSanction {
			status = sanction.TEMP_STATUS_UNSANCTIONED
		}
		rv = append(rv, &sanction.TemporaryEntry{
			Address:    addr.String(),
			ProposalId: id,
			Status:     status,
		})
		return false
	})
	return rv
}
