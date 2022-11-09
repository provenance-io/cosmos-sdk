package keeper_test

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

type TestSuite struct {
	suite.Suite

	app        *simapp.SimApp
	sdkCtx     sdk.Context
	stdlibCtx  context.Context
	keeper     keeper.Keeper
	bankKeeper bankkeeper.Keeper

	blockTime time.Time
	addr1     sdk.AccAddress
	addr2     sdk.AccAddress
	addr3     sdk.AccAddress
	addr4     sdk.AccAddress
	addr5     sdk.AccAddress
}

func (s *TestSuite) SetupTest() {
	s.blockTime = tmtime.Now()
	s.app = simapp.Setup(s.T(), false)
	s.sdkCtx = s.app.BaseApp.NewContext(false, tmproto.Header{}).WithBlockHeader(tmproto.Header{Time: s.blockTime})
	s.stdlibCtx = sdk.WrapSDKContext(s.sdkCtx)
	s.keeper = s.app.QuarantineKeeper
	s.bankKeeper = s.app.BankKeeper

	addrs := simapp.AddTestAddrsIncremental(s.app, s.sdkCtx, 5, sdk.NewInt(1_000_000_000))
	s.addr1 = addrs[0]
	s.addr2 = addrs[1]
	s.addr3 = addrs[2]
	s.addr4 = addrs[3]
	s.addr5 = addrs[4]
}

func (s *TestSuite) StopIfFailed() {
	if s.T().Failed() {
		s.T().FailNow()
	}
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
		s.keeper.SetFundsHolder(s.addr1)

		actual := s.keeper.GetFundsHolder()
		s.Assert().Equal(s.addr1, actual, "funds holder")
	})
}

func (s *TestSuite) TestQuarantineOptInOut() {
	s.Run("is quarantined before opting in", func() {
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in and check", func() {
		err := s.keeper.SetOptIn(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptIn addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().True(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in again and check", func() {
		err := s.keeper.SetOptIn(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptIn addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().True(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt out and check", func() {
		err := s.keeper.SetOptOut(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptOut addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt out again and check", func() {
		err := s.keeper.SetOptOut(s.sdkCtx, s.addr2)
		s.Require().NoError(err, "SetOptOut addr2")
		actual := s.keeper.IsQuarantinedAddr(s.sdkCtx, s.addr2)
		s.Assert().False(actual, "IsQuarantinedAddr addr2")
	})

	s.Run("opt in event", func() {
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		err := s.keeper.SetOptIn(ctx, s.addr3)
		s.Require().NoError(err, "SetOptIn addr3")

		events := ctx.EventManager().Events().ToABCIEvents()
		s.Require().GreaterOrEqual(len(events), 1, "number of events")
		event := events[0]
		s.Assert().Equal("cosmos.quarantine.v1beta1.EventOptIn", event.Type, "event type")
		s.Require().GreaterOrEqual(len(event.Attributes), 1, "number of event attributes")
		s.Assert().Equal("to_address", string(event.Attributes[0].Key), "attribute key")
		s.Assert().Equal(`"`+s.addr3.String()+`"`, string(event.Attributes[0].Value), "attribute value")
	})

	s.Run("opt out event", func() {
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		err := s.keeper.SetOptOut(ctx, s.addr3)
		s.Require().NoError(err, "SetOptOut addr3")

		events := ctx.EventManager().Events().ToABCIEvents()
		s.Require().GreaterOrEqual(len(events), 1, "number of events")
		event := events[0]
		s.Assert().Equal("cosmos.quarantine.v1beta1.EventOptOut", event.Type, "event type")
		s.Require().GreaterOrEqual(len(event.Attributes), 1, "number of event attributes")
		s.Assert().Equal("to_address", string(event.Attributes[0].Key), "attribute key")
		s.Assert().Equal(`"`+s.addr3.String()+`"`, string(event.Attributes[0].Value), "attribute value")
	})
}

func (s *TestSuite) TestQuarantinedAccountsIterateAndGetAll() {
	// Opt in all of them except addr4.
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr1), "SetOptIn addr1")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr2), "SetOptIn addr2")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr3), "SetOptIn addr3")
	s.Require().NoError(s.keeper.SetOptIn(s.sdkCtx, s.addr5), "SetOptIn addr5")

	// Now opt out addr2.
	s.Require().NoError(s.keeper.SetOptOut(s.sdkCtx, s.addr2), "SetOptOut addr2")

	expectedAddrs := []sdk.AccAddress{s.addr1, s.addr3, s.addr5}
	sort.Slice(expectedAddrs, func(i, j int) bool {
		return bytes.Compare(expectedAddrs[i], expectedAddrs[j]) < 0
	})

	s.Run("IterateQuarantinedAccounts", func() {
		addrs := make([]sdk.AccAddress, 0, len(expectedAddrs))
		callback := func(toAddr sdk.AccAddress) bool {
			addrs = append(addrs, toAddr)
			return false
		}

		testFunc := func() {
			s.keeper.IterateQuarantinedAccounts(s.sdkCtx, callback)
		}
		s.Require().NotPanics(testFunc, "IterateQuarantinedAccounts")
		s.Assert().Equal(expectedAddrs, addrs, "iterated addrs")
	})

	s.Run("GetAllQuarantinedAccounts", func() {
		expected := make([]string, len(expectedAddrs))
		for i, addr := range expectedAddrs {
			expected[i] = addr.String()
		}

		var actual []string
		testFunc := func() {
			actual = s.keeper.GetAllQuarantinedAccounts(s.sdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllQuarantinedAccounts")
		s.Assert().Equal(expected, actual, "GetAllQuarantinedAccounts")
	})
}

// TODO[1046]: GetAutoResponse
// TODO[1046]: IsAutoAccept
// TODO[1046]: IsAutoDecline
// TODO[1046]: SetAutoResponse
// TODO[1046]: IterateAutoResponses
// TODO[1046]: GetAllAutoResponseEntries

// TODO[1046]: GetQuarantineRecord
// TODO[1046]: GetQuarantineRecords
// TODO[1046]: SetQuarantineRecord
// TODO[1046]: AddQuarantinedCoins
// TODO[1046]: IterateQuarantineRecords
// TODO[1046]: GetAllQuarantinedFunds
// TODO[1046]: AcceptQuarantinedFunds
// TODO[1046]: DeclineQuarantinedFunds
// TODO[1046]: bzToQuarantineRecord

// TODO[1046]: getQuarantineRecordSuffixIndex
// TODO[1046]: setQuarantineRecordSuffixIndex
// TODO[1046]: addQuarantineRecordSuffixIndexes
// TODO[1046]: deleteQuarantineRecordSuffixIndexes
// TODO[1046]: getQuarantineRecordSuffixes
// TODO[1046]: bzToQuarantineRecordSuffixIndex
