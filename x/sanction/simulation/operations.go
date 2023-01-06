package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

const (
	OpWeightSanction            = "op_weight_sanction"
	OpWeightSanctionImmediate   = "op_weight_sanction_immediate"
	OpWeightUnsanction          = "op_weight_unsanction"
	OpWeightUnsanctionImmediate = "op_weight_unsanction_immediate"
	OpWeightUpdateParams        = "op_weight_update_params"

	DefaultWeightSanction            = 10
	DefaultWeightSanctionImmediate   = 10
	DefaultWeightUnsanction          = 10
	DefaultWeightUnsanctionImmediate = 10
	DefaultWeightUpdateParams        = 10
)

// WeightedOpsArgs holds all the args provided to WeightedOperations so that they can be passed on later more easily.
type WeightedOpsArgs struct {
	appParams simtypes.AppParams
	cdc       codec.JSONCodec
	ak        sanction.AccountKeeper
	bk        sanction.BankKeeper
	gk        sanction.GovKeeper
	sk        *keeper.Keeper
	appCdc    cdctypes.AnyUnpacker
}

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak sanction.AccountKeeper, bk sanction.BankKeeper, gk sanction.GovKeeper, sk keeper.Keeper, appCdc cdctypes.AnyUnpacker,
) simulation.WeightedOperations {
	args := &WeightedOpsArgs{
		appParams: appParams,
		cdc:       cdc,
		ak:        ak,
		bk:        bk,
		gk:        gk,
		sk:        &keeper.Keeper{},
		appCdc:    appCdc,
	}

	var (
		weightSanction            int
		weightSanctionImmediate   int
		weightUnsanction          int
		weightUnsanctionImmediate int
		weightUpdateParams        int
	)

	appParams.GetOrGenerate(cdc, OpWeightSanction, &weightSanction, nil,
		func(_ *rand.Rand) { weightSanction = DefaultWeightSanction })
	appParams.GetOrGenerate(cdc, OpWeightSanctionImmediate, &weightSanctionImmediate, nil,
		func(_ *rand.Rand) { weightSanctionImmediate = DefaultWeightSanctionImmediate })
	appParams.GetOrGenerate(cdc, OpWeightUnsanction, &weightUnsanction, nil,
		func(_ *rand.Rand) { weightUnsanction = DefaultWeightUnsanction })
	appParams.GetOrGenerate(cdc, OpWeightUnsanctionImmediate, &weightUnsanctionImmediate, nil,
		func(_ *rand.Rand) { weightUnsanctionImmediate = DefaultWeightUnsanctionImmediate })
	appParams.GetOrGenerate(cdc, OpWeightUpdateParams, &weightUpdateParams, nil,
		func(_ *rand.Rand) { weightUpdateParams = DefaultWeightUpdateParams })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(weightSanction, SimulateGovMsgSanction(args)),
		simulation.NewWeightedOperation(weightSanctionImmediate, SimulateGovMsgSanctionImmediate(args)),
		simulation.NewWeightedOperation(weightUnsanction, SimulateGovMsgUnsanction(args)),
		simulation.NewWeightedOperation(weightUnsanctionImmediate, SimulateGovMsgUnsanctionImmediate(args)),
		simulation.NewWeightedOperation(weightUpdateParams, SimulateGovMsgUpdateParams(args)),
	}
}

func SimulateGovMsgSanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// I'll need this in here:
		//   return simtypes.NoOpMsg(sanction.ModuleName, msgType, "fix me"), nil, err
		// Pick an account to send the message
		// Create the message for the gov prop.
		// Create the gov prop.
		// Pick a deposit that's less than the immediate sanction amount but more than the minimum required.
		// Generate tx
		// 		encCfg := simappparams.MakeTestEncodingConfig()
		//		tx, err := helpers.GenSignedMockTx(
		//			r,
		//			encCfg.TxConfig,
		//			[]sdk.Msg{msg},
		//			fees,
		//			helpers.DefaultGenTxGas,
		//			chainID,
		//			[]uint64{account.GetAccountNumber()},
		//			[]uint64{account.GetSequence()},
		//			acct.PrivKey,
		//		)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		//		}
		// Send the tx.
		//		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		//		}
		// Return what we've done.
		//		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
		// TODO[1046]: SimulateGovMsgSanction
		panic("not implemented")
	}
}

func SimulateGovMsgSanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// I'll need this in here:
		//   return simtypes.NoOpMsg(sanction.ModuleName, msgType, "fix me"), nil, err
		// Pick an account to send the message
		// Create the message for the gov prop.
		// Create the gov prop.
		// Pick a deposit that's more than the immediate sanction amount and also more than the minimum required.
		// Generate tx
		// 		encCfg := simappparams.MakeTestEncodingConfig()
		//		tx, err := helpers.GenSignedMockTx(
		//			r,
		//			encCfg.TxConfig,
		//			[]sdk.Msg{msg},
		//			fees,
		//			helpers.DefaultGenTxGas,
		//			chainID,
		//			[]uint64{account.GetAccountNumber()},
		//			[]uint64{account.GetSequence()},
		//			acct.PrivKey,
		//		)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		//		}
		// Send the tx.
		//		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		//		}
		// Return what we've done.
		//		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
		// TODO[1046]: SimulateGovMsgSanctionImmediate
		panic("not implemented")
	}
}

func SimulateGovMsgUnsanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// I'll need this in here:
		//   return simtypes.NoOpMsg(sanction.ModuleName, msgType, "fix me"), nil, err
		// Pick an account to send the message
		// Create the message for the gov prop.
		// Create the gov prop.
		// Pick a deposit that's less than the immediate unsanction amount but more than the minimum required.
		// Generate tx
		// 		encCfg := simappparams.MakeTestEncodingConfig()
		//		tx, err := helpers.GenSignedMockTx(
		//			r,
		//			encCfg.TxConfig,
		//			[]sdk.Msg{msg},
		//			fees,
		//			helpers.DefaultGenTxGas,
		//			chainID,
		//			[]uint64{account.GetAccountNumber()},
		//			[]uint64{account.GetSequence()},
		//			acct.PrivKey,
		//		)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		//		}
		// Send the tx.
		//		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		//		}
		// Return what we've done.
		//		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
		// TODO[1046]: SimulateGovMsgUnsanction
		panic("not implemented")
	}
}

func SimulateGovMsgUnsanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// I'll need this in here:
		//   return simtypes.NoOpMsg(sanction.ModuleName, msgType, "fix me"), nil, err
		// Pick an account to send the message
		// Create the message for the gov prop.
		// Create the gov prop.
		// Pick a deposit that's more than the immediate unsanction amount and also more than the minimum required.
		// Generate tx
		// 		encCfg := simappparams.MakeTestEncodingConfig()
		//		tx, err := helpers.GenSignedMockTx(
		//			r,
		//			encCfg.TxConfig,
		//			[]sdk.Msg{msg},
		//			fees,
		//			helpers.DefaultGenTxGas,
		//			chainID,
		//			[]uint64{account.GetAccountNumber()},
		//			[]uint64{account.GetSequence()},
		//			acct.PrivKey,
		//		)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		//		}
		// Send the tx.
		//		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		//		}
		// Return what we've done.
		//		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
		// TODO[1046]: SimulateGovMsgUnsanctionImmediate
		panic("not implemented")
	}
}

func SimulateGovMsgUpdateParams(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// I'll need this in here:
		//   return simtypes.NoOpMsg(sanction.ModuleName, msgType, "fix me"), nil, err
		// Pick an account to send the message
		// Create the message for the gov prop.
		// Create the gov prop.
		// Pick a deposit that's more than the immediate unsanction amount and also more than the minimum required.
		// Generate tx
		// 		encCfg := simappparams.MakeTestEncodingConfig()
		//		tx, err := helpers.GenSignedMockTx(
		//			r,
		//			encCfg.TxConfig,
		//			[]sdk.Msg{msg},
		//			fees,
		//			helpers.DefaultGenTxGas,
		//			chainID,
		//			[]uint64{account.GetAccountNumber()},
		//			[]uint64{account.GetSequence()},
		//			acct.PrivKey,
		//		)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to generate mock tx"), nil, err
		//		}
		// Send the tx.
		//		_, _, err = app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
		//		if err != nil {
		//			return simtypes.NoOpMsg(quarantine.ModuleName, msgType, "unable to deliver tx"), nil, err
		//		}
		// Return what we've done.
		//		return simtypes.NewOperationMsg(msg, true, "", codec.NewProtoCodec(encCfg.InterfaceRegistry)), nil, nil
		// TODO[1046]: SimulateGovMsgUpdateParams
		panic("not implemented")
	}
}
