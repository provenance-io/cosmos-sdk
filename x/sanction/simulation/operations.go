package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak sanction.AccountKeeper, bk sanction.BankKeeper, k keeper.Keeper, appCdc cdctypes.AnyUnpacker,
) simulation.WeightedOperations {
	// TODO[1046]: Implement WeightedOperations
	panic("not implemented")
}
