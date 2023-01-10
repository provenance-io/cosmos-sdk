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
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

type SimTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *simapp.SimApp
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}

func (s *SimTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T(), false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// getTestingAccounts creates testing accounts with a default balance.
func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, count int) []simtypes.Account {
	return s.getTestingAccountsWithPower(r, count, 200)
}

// getTestingAccountsWithPower creates new accounts with the specified power (coins amount).
func (s *SimTestSuite) getTestingAccountsWithPower(r *rand.Rand, count int, power int64) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, count)

	initAmt := sdk.TokensFromConsensusPower(power, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(bankutil.FundAccount(s.app.BankKeeper, s.ctx, account.Address, initCoins))
	}

	return accounts
}

// setSanctionParamsAboveGovDeposit looks up the x/gov min deposit and sets the
// sanction params to be larger by 5 (for sanction) and 10 (for unsanction).
// If there's no gov min dep, sets params to 5stake and 10stake respectively.
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
		sancParams.ImmediateUnsanctionMinDeposit = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
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
			s.T().Logf("operationMsg.Msg: %s", operationMsg.Msg)
			s.Assert().NoError(err, "op error")
			s.Assert().Equal(exp.weight, actual.Weight(), "op weight")
			s.Assert().True(operationMsg.OK, "op msg ok")
			s.Assert().Equal(exp.comment, operationMsg.Comment, "op msg comment")
			s.Assert().Equal("gov", operationMsg.Route, "op msg route")
			s.Assert().Equal(govPropType, operationMsg.Name, "op msg name")
			s.Assert().Len(futureOps, accountCount, "future ops")
			// Note: As of writing this, the content of operationMsg.Msg comes from MsgSubmitProposal.GetSignBytes.
			// But for some reason, it's also wrapped in '{"type":"{msg.Type}","value":"{msg.GetSignBytes}"}'.
			// The sign bytes are json, but the MsgSubmitProposal.Messages field's json marshals as just the value
			// instead of the Any that it is (i.e. there's no type_url). That makes it impossible to know from
			// that operationMsg.Msg field what type of messages are in the proposal Messages.
			// For this specific case, both MsgSanction and MsgUnsanction look exactly the same,
			// it's just: '{"addresses":[...]}'
			// So, long story short (too late), there's nothing worthwhile to check in the operationMsg.Msg field.
		})
	}
}

func (s *SimTestSuite) TestSendGovMsg() {
	r := rand.New(rand.NewSource(1))
	accounts := s.getTestingAccounts(r, 10)
	accounts = append(accounts, s.getTestingAccountsWithPower(r, 1, 0)...)
	accounts = append(accounts, s.getTestingAccountsWithPower(r, 1, 1)...)
	acctZero := accounts[len(accounts)-2]
	acctOne := accounts[len(accounts)-1]
	acctOneBalance := s.app.BankKeeper.SpendableCoins(s.ctx, acctOne.Address)
	var acctOneBalancePlusOne sdk.Coins
	for _, c := range acctOneBalance {
		acctOneBalancePlusOne = acctOneBalancePlusOne.Add(sdk.NewCoin(c.Denom, c.Amount.AddRaw(1)))
	}

	tests := []struct {
		name            string
		sender          simtypes.Account
		msg             sdk.Msg
		deposit         sdk.Coins
		comment         string
		expSkip         bool
		expOpMsgRoute   string
		expOpMsgName    string
		expOpMsgComment string
		expInErr        []string
	}{
		{
			name:   "no spendable coins",
			sender: acctZero,
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[4].Address.String(), accounts[5].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "should not matter",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has no spendable coins",
			expInErr:        nil,
		},
		{
			name:   "not enough coins for deposit",
			sender: acctOne,
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[5].Address.String(), accounts[6].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         acctOneBalancePlusOne,
			comment:         "should not be this",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&sanction.MsgSanction{}),
			expOpMsgComment: "sender has insufficient balance to cover deposit",
			expInErr:        nil,
		},
		{
			name:            "nil msg",
			sender:          accounts[0],
			msg:             nil,
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "will not get returned",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    "/",
			expOpMsgComment: "wrapping MsgSanction as Any",
			expInErr:        []string{"Expecting non nil value to create a new Any", "failed packing protobuf message to Any"},
		},
		{
			name: "gen and deliver returns error",
			sender: simtypes.Account{
				PrivKey: accounts[0].PrivKey,
				PubKey:  acctOne.PubKey,
				Address: acctOne.Address,
				ConsKey: accounts[0].ConsKey,
			},
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[6].Address.String(), accounts[7].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         acctOneBalance,
			comment:         "this should be ignored",
			expSkip:         true,
			expOpMsgRoute:   "sanction",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "unable to deliver tx",
			expInErr:        []string{"pubKey does not match signer address", "invalid pubkey"},
		},
		{
			name:   "all good",
			sender: accounts[1],
			msg: &sanction.MsgSanction{
				Addresses: []string{accounts[2].Address.String(), accounts[3].Address.String()},
				Authority: s.app.SanctionKeeper.GetAuthority(),
			},
			deposit:         sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 5)},
			comment:         "this is a test comment",
			expSkip:         false,
			expOpMsgRoute:   "gov",
			expOpMsgName:    sdk.MsgTypeURL(&govv1.MsgSubmitProposal{}),
			expOpMsgComment: "this is a test comment",
			expInErr:        nil,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			args := &simulation.SendGovMsgArgs{
				WeightedOpsArgs: simulation.WeightedOpsArgs{
					AppParams:  make(simtypes.AppParams),
					JSONCodec:  s.app.AppCodec(),
					ProtoCodec: codec.NewProtoCodec(s.app.InterfaceRegistry()),
					AK:         s.app.AccountKeeper,
					BK:         s.app.BankKeeper,
					GK:         s.app.GovKeeper,
					SK:         &s.app.SanctionKeeper,
				},
				R:       rand.New(rand.NewSource(1)),
				App:     s.app.BaseApp,
				Ctx:     s.ctx,
				Accs:    accounts,
				ChainID: "send-gov-test",
				Sender:  tc.sender,
				Msg:     tc.msg,
				Deposit: tc.deposit,
				Comment: tc.comment,
			}

			var skip bool
			var opMsg simtypes.OperationMsg
			var err error
			testFunc := func() {
				skip, opMsg, err = simulation.SendGovMsg(args)
			}
			s.Require().NotPanics(testFunc, "SendGovMsg")
			testutil.AssertErrorContents(s.T(), err, tc.expInErr, "SendGovMsg error")
			s.Assert().Equal(tc.expSkip, skip, "SendGovMsg result skip bool")
			s.Assert().Equal(tc.expOpMsgRoute, opMsg.Route, "SendGovMsg result op msg route")
			s.Assert().Equal(tc.expOpMsgName, opMsg.Name, "SendGovMsg result op msg name")
			s.Assert().Equal(tc.expOpMsgComment, opMsg.Comment, "SendGovMsg result op msg comment")
			if !tc.expSkip && !skip {
				// If we don't expect a skip, and we didn't get one,
				// get the last gov prop and make sure it's the one we just sent.
				expMsgs := []sdk.Msg{tc.msg}
				props := s.app.GovKeeper.GetProposals(s.ctx)
				if s.Assert().NotEmpty(props, "GovKeeper.GetProposals result should at least have the entry we just tried to create") {
					prop := props[len(props)-1]
					msgs, err := prop.GetMsgs()
					if s.Assert().NoError(err, "error from prop.GetMsgs() on the last gov prop") {
						s.Assert().Equal(expMsgs, msgs, "messages in the last gov prop")
					}
				}
			}
		})
	}
}

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
			name: "nil empty",
			a:    nil,
			b:    sdk.Coins{},
			exp:  sdk.Coins{},
		},
		{
			name: "empty nil",
			a:    sdk.Coins{},
			b:    nil,
			exp:  sdk.Coins{},
		},
		{
			name: "empty empty",
			a:    sdk.Coins{},
			b:    sdk.Coins{},
			exp:  sdk.Coins{},
		},
		{
			name: "one denom nil",
			a:    cz("5acoin"),
			b:    nil,
			exp:  cz("5acoin"),
		},
		{
			name: "one denom empty",
			a:    cz("5acoin"),
			b:    sdk.Coins{},
			exp:  cz("5acoin"),
		},
		{
			name: "nil one denom",
			a:    nil,
			b:    cz("3bcoin"),
			exp:  cz("3bcoin"),
		},
		{
			name: "empty one denom",
			a:    sdk.Coins{},
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
			name: "two denoms empty",
			a:    cz("1aone,2atwo"),
			b:    sdk.Coins{},
			exp:  cz("1aone,2atwo"),
		},
		{
			name: "nil two denoms",
			a:    nil,
			b:    cz("4bone,5btwo"),
			exp:  cz("4bone,5btwo"),
		},
		{
			name: "empty two denoms",
			a:    sdk.Coins{},
			b:    cz("4bone,5btwo"),
			exp:  cz("4bone,5btwo"),
		},
		{
			name: "different denoms",
			a:    cz("99acoin"),
			b:    cz("101bcoin"),
			exp:  cz("99acoin,101bcoin"),
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
		{
			name: "multiple denoms one shared a bigger",
			a:    cz("9aonlycoin,22sharecoin"),
			b:    cz("6bonlycoin,7bonlytwo,21sharecoin"),
			exp:  cz("9aonlycoin,6bonlycoin,7bonlytwo,22sharecoin"),
		},
		{
			name: "multiple denoms one shared b bigger",
			a:    cz("9aonlycoin,22sharecoin"),
			b:    cz("6bonlycoin,7bonlytwo,23sharecoin"),
			exp:  cz("9aonlycoin,6bonlycoin,7bonlytwo,23sharecoin"),
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
