package keeper_test

import (
	"context"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type TestSuite struct {
	suite.Suite

	app        *simapp.SimApp
	sdkCtx     sdk.Context
	ctx        context.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper

	blockTime time.Time
	addrs     []sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.blockTime = tmtime.Now()
	s.app = simapp.Setup(s.T(), false)
	s.sdkCtx = s.app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeader(tmproto.Header{Time: s.blockTime})
	s.ctx = sdk.WrapSDKContext(s.sdkCtx)
	s.keeper = s.app.QuarantineKeeper
	s.bankKeeper = s.app.BankKeeper

	s.addrs = simapp.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, sdk.NewInt(1_000_000_000))
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestGetFundsHolder() {
	// make sure it's back to normal after this test.
	orig := s.keeper.GetFundsHolder()
	defer s.keeper.SetFundsHolder(orig)

	s.Run("initial value", func() {
		expected := authtypes.NewModuleAddress(quarantine.ModuleName)

		actual := s.keeper.GetFundsHolder()
		s.Assert().Equal(expected, actual, "funds holder")
	})

	s.Run("set to nil", func() {
		s.keeper.SetFundsHolder(nil)

		actual := s.keeper.GetFundsHolder()
		s.Assert().Nil(actual, "funds holder")
	})

	s.Run("set to something else", func() {
		s.keeper.SetFundsHolder(s.addrs[0])

		actual := s.keeper.GetFundsHolder()
		s.Assert().Equal(s.addrs[0], actual, "funds holder")
	})
}

// TODO[1046]: IsQuarantinedAddr
// TODO[1046]: SetOptIn
// TODO[1046]: SetOptOut
// TODO[1046]: IterateQuarantinedAccounts
// TODO[1046]: GetAllQuarantinedAccounts
// TODO[1046]: GetAutoResponse
// TODO[1046]: IsAutoAccept
// TODO[1046]: IsAutoDecline
// TODO[1046]: SetAutoResponse
// TODO[1046]: IterateAutoResponses
// TODO[1046]: GetAllAutoResponseEntries
// TODO[1046]: GetQuarantineRecord
// TODO[1046]: SetQuarantineRecord
// TODO[1046]: AddQuarantinedCoins
// TODO[1046]: IterateQuarantineRecords
// TODO[1046]: GetAllQuarantinedFunds
// TODO[1046]: SetQuarantineRecordAccepted
// TODO[1046]: SetQuarantineRecordDeclined
// TODO[1046]: bzToQuarantineRecord
