package simulation

import (
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
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

// sendGovMsgArgs holds all the args available and needed for sending a gov msg.
type sendGovMsgArgs struct {
	r       *rand.Rand
	app     *baseapp.BaseApp
	ctx     sdk.Context
	accs    []simtypes.Account
	chainID string

	sender  simtypes.Account
	msg     sdk.Msg
	deposit sdk.Coins
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

// sendGovMsg sends a msg as a gov prop.
// It returns whether to skip the rest, an operation message, and any error encountered.
func sendGovMsg(wopArgs *WeightedOpsArgs, args *sendGovMsgArgs) (bool, simtypes.OperationMsg, error) {
	msgType := sdk.MsgTypeURL(args.msg)
	msgAny, err := codectypes.NewAnyWithValue(args.msg)
	if err != nil {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "wrapping MsgSanction as Any"), err
	}
	govMsg := &govv1.MsgSubmitProposal{
		Messages:       []*codectypes.Any{msgAny},
		InitialDeposit: args.deposit,
		Proposer:       args.sender.Address.String(),
		Metadata:       "",
	}

	spendableCoins := wopArgs.bk.SpendableCoins(args.ctx, args.sender.Address)

	if spendableCoins.Empty() {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "unable to generate fees"), nil
	}

	depDenomIndex := args.r.Intn(len(args.deposit))
	deposit := args.deposit[depDenomIndex]
	_, hasNeg := spendableCoins.SafeSub(deposit)
	if hasNeg {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "insufficient denom of deposit"), nil
	}

	senderAccount := wopArgs.ak.GetAccount(args.ctx, args.sender.Address)

	encCfg := simappparams.MakeTestEncodingConfig()

	tx, err := helpers.GenSignedMockTx(
		args.r,
		encCfg.TxConfig,
		[]sdk.Msg{govMsg},
		sdk.Coins{deposit},
		helpers.DefaultGenTxGas,
		args.chainID,
		[]uint64{senderAccount.GetAccountNumber()},
		[]uint64{senderAccount.GetSequence()},
		args.sender.PrivKey,
	)

	_, _, err = args.app.SimDeliver(encCfg.TxConfig.TxEncoder(), tx)
	if err != nil {
		return true, simtypes.NoOpMsg(sanction.ModuleName, msgType, "unable to deliver tx"), err
	}

	return false, simtypes.NewOperationMsg(args.msg, true, "", nil), nil
}

// operationMsgVote returns an operation that casts a yes vote on a gov prop from an account.
func operationMsgVote(args *WeightedOpsArgs, simAccount simtypes.Account, govPropID uint64, vote govv1.VoteOption) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := govv1.NewMsgVote(simAccount.Address, govPropID, vote, "")

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: sdk.Coins{},
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   args.ak,
			Bankkeeper:      args.bk,
			ModuleName:      sanction.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// maxCoins combines a and b taking the max of each denom.
// The result will have all the denoms from a and all the denoms from b.
// The amount of each denom is the max between a and b for that denom.
func maxCoins(a, b sdk.Coins) sdk.Coins {
	allDenomsMap := map[string]bool{}
	for _, c := range a {
		allDenomsMap[c.Denom] = true
	}
	for _, c := range a {
		allDenomsMap[c.Denom] = true
	}
	rv := make([]sdk.Coin, 0, len(allDenomsMap))
	for denom := range allDenomsMap {
		cA := a.AmountOf(denom)
		cB := a.AmountOf(denom)
		if cA.GT(cB) {
			rv = append(rv, sdk.NewCoin(denom, cA))
		} else {
			rv = append(rv, sdk.NewCoin(denom, cB))
		}
	}
	return sdk.NewCoins(rv...)
}

func SimulateGovMsgSanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgSanction{
			Authority: args.sk.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// First, get the governance min deposit needed and immediate sanction min deposit needed.
		govMinDep := sdk.NewCoins(args.gk.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.sk.GetImmediateSanctionMinDeposit(ctx)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot sanction without it being immediate"), nil, nil
		}

		sender, senderI := simtypes.RandomAcc(r, accs)
		// Create 1-10 new accounts to use.
		acctsToUse := simtypes.RandomAccounts(r, r.Intn(10)+1)
		// If there are 20 or more accounts, pick a random one for each 20.
		if len(accs) >= 20 {
			valsUsed := map[int]bool{senderI: true}
			for i := 0; i < len(accs)/20; i++ {
				for {
					acct, v := simtypes.RandomAcc(r, accs)
					if !valsUsed[v] {
						valsUsed[v] = true
						acctsToUse = append(acctsToUse, acct)
						break
					}
				}
			}
		}
		for _, acct := range acctsToUse {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		skip, opMsg, err := sendGovMsg(args, &sendGovMsgArgs{
			r:       r,
			app:     app,
			ctx:     ctx,
			accs:    accs,
			chainID: chainID,
			sender:  sender,
			msg:     msg,
			deposit: govMinDep,
		})

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.gk.GetProposalID(ctx)

		votingPeriod := args.gk.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationMsgVote(args, acct, proposalID, govv1.OptionYes),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgSanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgSanction{
			Authority: args.sk.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// Make sure an immediate sanction is possible.
		imMinDep := args.sk.GetImmediateSanctionMinDeposit(ctx)
		if imMinDep.IsZero() {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "immediate sanction min deposit is zero"), nil, nil
		}

		// Get the governance min deposit needed.
		govMinDep := sdk.NewCoins(args.gk.GetDepositParams(ctx).MinDeposit...)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot sanction without it being immediate"), nil, nil
		}

		deposit := maxCoins(imMinDep, govMinDep)

		sender, senderI := simtypes.RandomAcc(r, accs)
		// Create 1-10 new accounts to use.
		acctsToUse := simtypes.RandomAccounts(r, r.Intn(10)+1)
		// If there are 20 or more accounts, pick a random one for each 20.
		if len(accs) >= 20 {
			valsUsed := map[int]bool{senderI: true}
			for i := 0; i < len(accs)/20; i++ {
				for {
					acct, v := simtypes.RandomAcc(r, accs)
					if !valsUsed[v] {
						valsUsed[v] = true
						acctsToUse = append(acctsToUse, acct)
						break
					}
				}
			}
		}
		for _, acct := range acctsToUse {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		skip, opMsg, err := sendGovMsg(args, &sendGovMsgArgs{
			r:       r,
			app:     app,
			ctx:     ctx,
			accs:    accs,
			chainID: chainID,
			sender:  sender,
			msg:     msg,
			deposit: deposit,
		})

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.gk.GetProposalID(ctx)

		vote := govv1.OptionYes
		if r.Intn(2) == 0 {
			vote = govv1.OptionNo
		}

		votingPeriod := args.gk.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationMsgVote(args, acct, proposalID, vote),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUnsanction(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgUnsanction{
			Authority: args.sk.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// First, get the governance min deposit needed and immediate sanction min deposit needed.
		govMinDep := sdk.NewCoins(args.gk.GetDepositParams(ctx).MinDeposit...)
		imMinDep := args.sk.GetImmediateUnsanctionMinDeposit(ctx)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot unsanction without it being immediate"), nil, nil
		}

		sender, senderI := simtypes.RandomAcc(r, accs)
		// Create 1-10 new accounts to use.
		acctsToUse := simtypes.RandomAccounts(r, r.Intn(10)+1)
		// If there are 20 or more accounts, pick a random one for each 20.
		if len(accs) >= 20 {
			valsUsed := map[int]bool{senderI: true}
			for i := 0; i < len(accs)/20; i++ {
				for {
					acct, v := simtypes.RandomAcc(r, accs)
					if !valsUsed[v] {
						valsUsed[v] = true
						acctsToUse = append(acctsToUse, acct)
						break
					}
				}
			}
		}
		for _, acct := range acctsToUse {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		skip, opMsg, err := sendGovMsg(args, &sendGovMsgArgs{
			r:       r,
			app:     app,
			ctx:     ctx,
			accs:    accs,
			chainID: chainID,
			sender:  sender,
			msg:     msg,
			deposit: govMinDep,
		})

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.gk.GetProposalID(ctx)

		votingPeriod := args.gk.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationMsgVote(args, acct, proposalID, govv1.OptionYes),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUnsanctionImmediate(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msg := &sanction.MsgUnsanction{
			Authority: args.sk.GetAuthority(),
		}
		msgType := sdk.MsgTypeURL(msg)

		// Make sure an immediate unsanction is possible.
		imMinDep := args.sk.GetImmediateUnsanctionMinDeposit(ctx)
		if imMinDep.IsZero() {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "immediate unsanction min deposit is zero"), nil, nil
		}

		// Get the governance min deposit needed.
		govMinDep := sdk.NewCoins(args.gk.GetDepositParams(ctx).MinDeposit...)
		if !imMinDep.IsZero() && govMinDep.IsAllGTE(imMinDep) {
			return simtypes.NoOpMsg(sanction.ModuleName, msgType, "cannot unsanction without it being immediate"), nil, nil
		}

		deposit := maxCoins(imMinDep, govMinDep)

		sender, senderI := simtypes.RandomAcc(r, accs)
		// Create 1-10 new accounts to use.
		acctsToUse := simtypes.RandomAccounts(r, r.Intn(10)+1)
		// If there are 20 or more accounts, pick a random one for each 20.
		if len(accs) >= 20 {
			valsUsed := map[int]bool{senderI: true}
			for i := 0; i < len(accs)/20; i++ {
				for {
					acct, v := simtypes.RandomAcc(r, accs)
					if !valsUsed[v] {
						valsUsed[v] = true
						acctsToUse = append(acctsToUse, acct)
						break
					}
				}
			}
		}
		for _, acct := range acctsToUse {
			msg.Addresses = append(msg.Addresses, acct.Address.String())
		}

		skip, opMsg, err := sendGovMsg(args, &sendGovMsgArgs{
			r:       r,
			app:     app,
			ctx:     ctx,
			accs:    accs,
			chainID: chainID,
			sender:  sender,
			msg:     msg,
			deposit: deposit,
		})

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.gk.GetProposalID(ctx)

		vote := govv1.OptionYes
		if r.Intn(2) == 0 {
			vote = govv1.OptionNo
		}

		votingPeriod := args.gk.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationMsgVote(args, acct, proposalID, vote),
			}
		}

		return opMsg, fops, nil
	}
}

func SimulateGovMsgUpdateParams(args *WeightedOpsArgs) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get the governance min deposit needed.
		govMinDep := sdk.NewCoins(args.gk.GetDepositParams(ctx).MinDeposit...)

		sender, _ := simtypes.RandomAcc(r, accs)

		msg := &sanction.MsgUpdateParams{
			Params:    RandomParams(r),
			Authority: args.sk.GetAuthority(),
		}

		skip, opMsg, err := sendGovMsg(args, &sendGovMsgArgs{
			r:       r,
			app:     app,
			ctx:     ctx,
			accs:    accs,
			chainID: chainID,
			sender:  sender,
			msg:     msg,
			deposit: govMinDep,
		})

		if skip || err != nil {
			return opMsg, nil, err
		}

		proposalID, err := args.gk.GetProposalID(ctx)

		votingPeriod := args.gk.GetVotingParams(ctx).VotingPeriod
		fops := make([]simtypes.FutureOperation, len(accs))
		for i, acct := range accs {
			whenVote := ctx.BlockHeader().Time.Add(time.Duration(r.Int63n(int64(votingPeriod.Seconds()))) * time.Second)
			fops[i] = simtypes.FutureOperation{
				BlockTime: whenVote,
				Op:        operationMsgVote(args, acct, proposalID, govv1.OptionYes),
			}
		}

		return opMsg, fops, nil
	}
}
