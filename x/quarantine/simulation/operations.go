package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak quarantine.AccountKeeper,
	bk quarantine.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker,
) simulation.WeightedOperations {
	// TODO[1046]: Implement WeightedOperations and all the operations.
	return nil
}
