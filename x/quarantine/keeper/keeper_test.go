package keeper_test

import (
	"bytes"
	"context"
	"fmt"
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

func (s *TestSuite) TestAutoResponseGettersSetter() {
	addrs := []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5}
	allResps := []quarantine.AutoResponse{
		quarantine.AUTO_RESPONSE_ACCEPT,
		quarantine.AUTO_RESPONSE_DECLINE,
		quarantine.AUTO_RESPONSE_UNSPECIFIED,
	}

	s.Run("GetAutoResponse on unset addrs", func() {
		expected := quarantine.AUTO_RESPONSE_UNSPECIFIED
		for i, addrI := range addrs {
			for j, addrJ := range addrs {
				if i == j {
					continue
				}
				actual := s.keeper.GetAutoResponse(s.sdkCtx, addrI, addrJ)
				s.Assert().Equal(expected, actual, "GetAutoResponse addr%d addr%d", i+1, j+1)
			}
		}
	})

	s.Run("GetAutoResponse on same addr", func() {
		expected := quarantine.AUTO_RESPONSE_ACCEPT
		for i, addr := range addrs {
			actual := s.keeper.GetAutoResponse(s.sdkCtx, addr, addr)
			s.Assert().Equal(expected, actual, "GetAutoResponse addr%d addr%d", i+1, i+1)
		}
	})

	for _, expected := range allResps {
		s.Run(fmt.Sprintf("set %s", expected), func() {
			testFunc := func() {
				s.keeper.SetAutoResponse(s.sdkCtx, s.addr3, s.addr1, expected)
			}
			s.Require().NotPanics(testFunc, "SetAutoResponse addr3 addr1 %s", expected)
			actual := s.keeper.GetAutoResponse(s.sdkCtx, s.addr3, s.addr1)
			s.Assert().Equal(expected, actual, "GetAutoResponse after set %s", expected)
		})
	}

	s.Run("IsAutoAccept", func() {
		testFunc := func() {
			s.keeper.SetAutoResponse(s.sdkCtx, s.addr4, s.addr2, quarantine.AUTO_RESPONSE_ACCEPT)
		}
		s.Require().NotPanics(testFunc, "SetAutoResponse")

		actual42 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr2)
		s.Assert().True(actual42, "IsAutoAccept addr4 addr2")
		actual43 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr3)
		s.Assert().False(actual43, "IsAutoAccept addr4 addr3")
		actual44 := s.keeper.IsAutoAccept(s.sdkCtx, s.addr4, s.addr4)
		s.Assert().True(actual44, "IsAutoAccept self")
	})

	s.Run("IsAutoDecline", func() {
		testFunc := func() {
			s.keeper.SetAutoResponse(s.sdkCtx, s.addr5, s.addr2, quarantine.AUTO_RESPONSE_DECLINE)
		}
		s.Require().NotPanics(testFunc, "SetAutoResponse")

		actual52 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr2)
		s.Assert().True(actual52, "IsAutoDecline addr5 addr2")
		actual53 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr3)
		s.Assert().False(actual53, "IsAutoDecline addr5 addr3")
		actual55 := s.keeper.IsAutoDecline(s.sdkCtx, s.addr5, s.addr5)
		s.Assert().False(actual55, "IsAutoDecline self")
	})
}

func (s *TestSuite) TestAutoResponsesItateAndGetAll() {
	setAutoTestFunc := func(addrA, addrB sdk.AccAddress, response quarantine.AutoResponse) func() {
		return func() {
			s.keeper.SetAutoResponse(s.sdkCtx, addrA, addrB, response)
		}
	}
	// Shorten up the names a bit.
	arAccept := quarantine.AUTO_RESPONSE_ACCEPT
	arDecline := quarantine.AUTO_RESPONSE_DECLINE
	arUnspecified := quarantine.AUTO_RESPONSE_UNSPECIFIED

	// Set up some auto-responses.
	// This is purposely done in a random order.

	// Set account 1 to auto-accept from all.
	// Set account 2 to auto-accept from all.
	// Set 3 to auto-decline from all
	// Set 4 to auto-accept from 2 and 3 and auto-decline from 5
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr3, arAccept), "4 <- 3 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr2, arDecline), "3 <- 2 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr5, arAccept), "2 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr5, arDecline), "3 <- 5 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr4, arAccept), "2 <- 4 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr1, arAccept), "2 <- 1 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr5, arDecline), "4 <- 5 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr4, arDecline), "3 <- 4 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr1, arDecline), "3 <- 1 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr5, arAccept), "1 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr2, arAccept), "1 <- 2 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr4, arAccept), "1 <- 4 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr3, arAccept), "2 <- 3 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr2, arAccept), "4 <- 2 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr1, s.addr3, arAccept), "1 <- 3 accept")

	// Now undo/change a few of those.
	// Set 2 to unspecified from 3 and 4
	// Set 3 to unspecified from 5
	// Set 4 to auto-decline from 3 and auto-accept from 5
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr5, arAccept), "4 <- 5 accept")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr3, arUnspecified), "2 <- 3 unspecified")
	s.Require().NotPanics(setAutoTestFunc(s.addr3, s.addr5, arUnspecified), "3 <- 5 unspecified")
	s.Require().NotPanics(setAutoTestFunc(s.addr4, s.addr3, arDecline), "4 <- 3 decline")
	s.Require().NotPanics(setAutoTestFunc(s.addr2, s.addr4, arUnspecified), "2 <- 4 unspecified")

	// Setup result:
	// 1 <- 2 = accept   3 <- 1 = decline
	// 1 <- 3 = accept   3 <- 2 = decline
	// 1 <- 4 = accept   3 <- 4 = decline
	// 1 <- 5 = accept   4 <- 2 = accept
	// 2 <- 1 = accept   4 <- 3 = decline
	// 2 <- 5 = accept   4 <- 5 = accept

	// Let's hope the addresses are actually incremental or else this gets a lot tougher to define.
	type callbackArgs struct {
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		response quarantine.AutoResponse
	}

	expectedAllArgs := []callbackArgs{
		{toAddr: s.addr1, fromAddr: s.addr2, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr3, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr4, response: arAccept},
		{toAddr: s.addr1, fromAddr: s.addr5, response: arAccept},
		{toAddr: s.addr2, fromAddr: s.addr1, response: arAccept},
		{toAddr: s.addr2, fromAddr: s.addr5, response: arAccept},
		{toAddr: s.addr3, fromAddr: s.addr1, response: arDecline},
		{toAddr: s.addr3, fromAddr: s.addr2, response: arDecline},
		{toAddr: s.addr3, fromAddr: s.addr4, response: arDecline},
		{toAddr: s.addr4, fromAddr: s.addr2, response: arAccept},
		{toAddr: s.addr4, fromAddr: s.addr3, response: arDecline},
		{toAddr: s.addr4, fromAddr: s.addr5, response: arAccept},
	}

	s.Run("IterateAutoResponses all", func() {
		actualAllArgs := make([]callbackArgs, 0, len(expectedAllArgs))
		callback := func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) bool {
			actualAllArgs = append(actualAllArgs, callbackArgs{toAddr: toAddr, fromAddr: fromAddr, response: response})
			return false
		}
		testFunc := func() {
			s.keeper.IterateAutoResponses(s.sdkCtx, nil, callback)
		}
		s.Require().NotPanics(testFunc, "IterateAutoResponses")
		s.Assert().Equal(expectedAllArgs, actualAllArgs, "iterated args")
	})

	for i, addr := range []sdk.AccAddress{s.addr1, s.addr2, s.addr3, s.addr4, s.addr5} {
		s.Run(fmt.Sprintf("IterateAutoResponses addr%d", i+1), func() {
			var expected []callbackArgs
			for _, args := range expectedAllArgs {
				if addr.Equals(args.toAddr) {
					expected = append(expected, args)
				}
			}
			var actual []callbackArgs
			callback := func(toAddr, fromAddr sdk.AccAddress, response quarantine.AutoResponse) bool {
				actual = append(actual, callbackArgs{toAddr: toAddr, fromAddr: fromAddr, response: response})
				return false
			}
			testFunc := func() {
				s.keeper.IterateAutoResponses(s.sdkCtx, addr, callback)
			}
			s.Require().NotPanics(testFunc, "IterateAutoResponses")
			s.Assert().Equal(expected, actual, "iterated args")
		})
	}

	s.Run("GetAllAutoResponseEntries", func() {
		expected := make([]*quarantine.AutoResponseEntry, len(expectedAllArgs))
		for i, args := range expectedAllArgs {
			expected[i] = &quarantine.AutoResponseEntry{
				ToAddress:   args.toAddr.String(),
				FromAddress: args.fromAddr.String(),
				Response:    args.response,
			}
		}

		var actual []*quarantine.AutoResponseEntry
		testFunc := func() {
			actual = s.keeper.GetAllAutoResponseEntries(s.sdkCtx)
		}
		s.Require().NotPanics(testFunc, "GetAllAutoResponseEntries")
		s.Assert().Equal(expected, actual, "GetAllAutoResponseEntries results")
	})
}

// TODO[1046]: SetQuarantineRecord
// TODO[1046]: bzToQuarantineRecord
// TODO[1046]: GetQuarantineRecord
// TODO[1046]: GetQuarantineRecords
// TODO[1046]: AddQuarantinedCoins
// TODO[1046]: AcceptQuarantinedFunds
// TODO[1046]: DeclineQuarantinedFunds
// TODO[1046]: IterateQuarantineRecords
// TODO[1046]: GetAllQuarantinedFunds

// TODO[1046]: setQuarantineRecordSuffixIndex
// TODO[1046]: bzToQuarantineRecordSuffixIndex
// TODO[1046]: getQuarantineRecordSuffixIndex
// TODO[1046]: getQuarantineRecordSuffixes
// TODO[1046]: addQuarantineRecordSuffixIndexes
// TODO[1046]: deleteQuarantineRecordSuffixIndexes
