package simulation_test

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(testutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins))
	}

	return accounts
}

func (s *SimTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T(), false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

func (s *SimTestSuite) setSanctionParamsAboveGovDeposit() {
	sancParams := &sanction.Params{
		ImmediateSanctionMinDeposit:   nil,
		ImmediateUnsanctionMinDeposit: nil,
	}

	for _, coin := range s.app.GovKeeper.GetDepositParams(s.ctx).MinDeposit {
		sanctCoin := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(5))
		unsanctCoin := sdk.NewCoin(coin.Denom, coin.Amount.AddRaw(10))
		sancParams.ImmediateSanctionMinDeposit = sancParams.ImmediateSanctionMinDeposit.Add(sanctCoin)
		sancParams.ImmediateUnsanctionMinDeposit = sancParams.ImmediateUnsanctionMinDeposit.Add(unsanctCoin)
	}

	if sancParams.ImmediateSanctionMinDeposit.IsZero() {
		sancParams.ImmediateSanctionMinDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}
	}
	if sancParams.ImmediateUnsanctionMinDeposit.IsZero() {
		sancParams.ImmediateUnsanctionMinDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)}
	}

	s.Require().NoError(s.app.SanctionKeeper.SetParams(s.ctx, sancParams), "SanctionKeeper.SetParams")
}

func (s *SimTestSuite) TestWeightedOperations() {
	s.setSanctionParamsAboveGovDeposit()

	govPropType := sdk.MsgTypeURL(&govv1.MsgSubmitProposal{})

	expected := []struct {
		comment string
		weight  int
	}{
		{comment: "sanction", weight: simulation.DefaultWeightSanction},
		{comment: "immediate sanction", weight: simulation.DefaultWeightSanctionImmediate},
		{comment: "unsanction", weight: simulation.DefaultWeightUnsanction},
		{comment: "immediate unsanction", weight: simulation.DefaultWeightUnsanctionImmediate},
		{comment: "update params", weight: simulation.DefaultWeightUpdateParams},
	}

	weightedOps := simulation.WeightedOperations(
		make(simtypes.AppParams), s.app.AppCodec(), codec.NewProtoCodec(s.app.InterfaceRegistry()),
		s.app.AccountKeeper, s.app.BankKeeper, s.app.GovKeeper, s.app.SanctionKeeper,
	)

	s.Require().Len(weightedOps, len(expected), "weighted ops")

	accountCount := 10
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, accountCount)

	for i, actual := range weightedOps {
		exp := expected[i]
		s.Run(exp.comment, func() {
			var operationMsg simtypes.OperationMsg
			var futureOps []simtypes.FutureOperation
			var err error
			testFunc := func() {
				operationMsg, futureOps, err = actual.Op()(r, s.app.BaseApp, s.ctx, accs, "")
			}
			s.Require().NotPanics(testFunc, "calling op")
			s.Assert().NoError(err, "op error")
			s.Assert().Equal(exp.weight, actual.Weight(), "op weight")
			s.Assert().True(operationMsg.OK, "op msg ok")
			s.Assert().Equal(exp.comment, operationMsg.Comment, "op msg comment")
			s.Assert().Equal("gov", operationMsg.Route, "op msg route")
			s.Assert().Equal(govPropType, operationMsg.Name, "op msg name")
			s.Assert().Len(futureOps, accountCount, "future ops")
		})
	}
}

// TODO[1046]: SendGovMsg
// TODO[1046]: OperationMsgVote

func TestMaxCoins(t *testing.T) {
	// Not using SimTestSuite for this one since it doesn't need the infrastructure.

	// cz is a short way to convert a string to Coins.
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name string
		a    sdk.Coins
		b    sdk.Coins
		exp  sdk.Coins
	}{
		{
			name: "nil nil",
			a:    nil,
			b:    nil,
			exp:  sdk.Coins{},
		},
		{
			name: "one denom nil",
			a:    cz("5acoin"),
			b:    nil,
			exp:  cz("5acoin"),
		},
		{
			name: "nil one denom",
			a:    nil,
			b:    cz("3bcoin"),
			exp:  cz("3bcoin"),
		},
		{
			name: "two denoms nil",
			a:    cz("1aone,2atwo"),
			b:    nil,
			exp:  cz("1aone,2atwo"),
		},
		{
			name: "nil two denoms",
			a:    nil,
			b:    cz("4bone,5btwo"),
			exp:  cz("4bone,5btwo"),
		},
		{
			name: "both have same denom a bigger",
			a:    cz("2sharecoin"),
			b:    cz("1sharecoin"),
			exp:  cz("2sharecoin"),
		},
		{
			name: "both have same denom b bigger",
			a:    cz("4sharecoin"),
			b:    cz("5sharecoin"),
			exp:  cz("5sharecoin"),
		},
		{
			name: "each with unique denoms",
			a:    cz("3aonecoin,8atwocoin"),
			b:    cz("4bonecoin,9btwocoin"),
			exp:  cz("3aonecoin,8atwocoin,4bonecoin,9btwocoin"),
		},
		{
			name: "each has multiple denoms only one is common a bigger",
			a:    cz("9aonlycoin,22sharecoin"),
			b:    cz("6bonlycoin,21sharecoin,7bonlytwo"),
			exp:  cz("9aonlycoin,6bonlycoin,7bonlytwo,22sharecoin"),
		},
		{
			name: "one denom smaller vs two denoms",
			a:    cz("1share"),
			b:    cz("2bcoin,2share"),
			exp:  cz("2bcoin,2share"),
		},
		{
			name: "one denom larger vs two denoms",
			a:    cz("3share"),
			b:    cz("2bcoin,2share"),
			exp:  cz("2bcoin,3share"),
		},
		{
			name: "two denoms vs one denom smaller",
			a:    cz("2acoin,2share"),
			b:    cz("1share"),
			exp:  cz("2acoin,2share"),
		},
		{
			name: "two denoms vs one denom larger",
			a:    cz("2acoin,2share"),
			b:    cz("3share"),
			exp:  cz("2acoin,3share"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.Coins
			testFunc := func() {
				actual = simulation.MaxCoins(tc.a, tc.b)
			}
			require.NotPanics(t, testFunc, "MaxCoins")
			assert.Equal(t, tc.exp.String(), actual.String(), "MaxCoins result")
		})
	}
}

// TODO[1046]: SimulateGovMsgSanction
// TODO[1046]: SimulateGovMsgSanctionImmediate
// TODO[1046]: SimulateGovMsgUnsanction
// TODO[1046]: SimulateGovMsgUnsanctionImmediate
// TODO[1046]: SimulateGovMsgUpdateParams
