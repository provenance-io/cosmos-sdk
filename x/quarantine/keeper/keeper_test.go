package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	. "github.com/cosmos/cosmos-sdk/x/quarantine/testutil"
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
	s.Run("initial value", func() {
		expected := authtypes.NewModuleAddress(quarantine.ModuleName)

		actual := s.keeper.GetFundsHolder()
		s.Assert().Equal(expected, actual, "funds holder")
	})

	s.Run("set to nil", func() {
		k := s.keeper.WithFundsHolder(nil)

		actual := k.GetFundsHolder()
		s.Assert().Nil(actual, "funds holder")
	})

	s.Run("set to something else", func() {
		k := s.keeper.WithFundsHolder(s.addr1)

		actual := k.GetFundsHolder()
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

		expected := sdk.Events{
			{
				Type: "cosmos.quarantine.v1beta1.EventOptIn",
				Attributes: []abci.EventAttribute{
					{
						Key:   []byte("to_address"),
						Value: []byte(fmt.Sprintf(`"%s"`, s.addr3.String())),
					},
				},
			},
		}
		actual := ctx.EventManager().Events()
		s.Assert().Equal(expected, actual, "emitted events")
	})

	s.Run("opt out event", func() {
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		err := s.keeper.SetOptOut(ctx, s.addr3)
		s.Require().NoError(err, "SetOptOut addr3")

		expected := sdk.Events{
			{
				Type: "cosmos.quarantine.v1beta1.EventOptOut",
				Attributes: []abci.EventAttribute{
					{
						Key:   []byte("to_address"),
						Value: []byte(fmt.Sprintf(`"%s"`, s.addr3.String())),
					},
				},
			},
		}
		actual := ctx.EventManager().Events()
		s.Assert().Equal(expected, actual, "emitted events")
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

func (s *TestSuite) TestBzToQuarantineRecord() {
	cdc := s.keeper.GetCodec()

	tests := []struct {
		name     string
		bz       []byte
		expected *quarantine.QuarantineRecord
		expErr   string
	}{
		{
			name: "control",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			},
		},
		{
			name: "nil bz",
			bz:   nil,
			expected: &quarantine.QuarantineRecord{
				Coins: sdk.Coins{},
			},
		},
		{
			name: "empty bz",
			bz:   nil,
			expected: &quarantine.QuarantineRecord{
				Coins: sdk.Coins{},
			},
		},
		{
			name: "not a quarantine record",
			bz: cdc.MustMarshal(&quarantine.AutoResponseEntry{
				ToAddress:   s.addr4.String(),
				FromAddress: s.addr3.String(),
				Response:    quarantine.AUTO_RESPONSE_ACCEPT,
			}),
			expErr: "proto: wrong wireType = 0 for field Coins",
		},
		{
			name:   "unknown bytes",
			bz:     []byte{0x75, 110, 0153, 0x6e, 0157, 119, 0156, 0xff, 0142, 0x79, 116, 0x65, 0163},
			expErr: "proto: illegal wireType 7",
		},
		{
			name: "declined",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                true,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                true,
			},
		},
		{
			name: "no unaccepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2, s.addr1, s.addr3},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2, s.addr1, s.addr3},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			},
		},
		{
			name: "no accepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr4, s.addr2, s.addr5},
				AcceptedFromAddresses:   []sdk.AccAddress{},
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr4, s.addr2, s.addr5},
				AcceptedFromAddresses:   nil,
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 888), sdk.NewInt64Coin("bar", 9000)),
				Declined:                false,
			},
		},
		{
			name: "no coins",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   nil,
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   sdk.Coins{},
				Declined:                false,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual *quarantine.QuarantineRecord
			var err error
			testFunc := func() {
				actual, err = s.keeper.BzToQuarantineRecord(tc.bz)
			}
			s.Require().NotPanics(testFunc, "bzToQuarantineRecord")
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "bzToQuarantineRecord error: %v", actual)
			} else {
				s.Require().NoError(err, "bzToQuarantineRecord error")
				s.Assert().Equal(tc.expected, actual, "bzToQuarantineRecord record")
			}
		})

		s.Run("must "+tc.name, func() {
			var actual *quarantine.QuarantineRecord
			testFunc := func() {
				actual = s.keeper.MustBzToQuarantineRecord(tc.bz)
			}
			if len(tc.expErr) > 0 {
				s.Require().PanicsWithError(tc.expErr, testFunc, "mustBzToQuarantineRecord: %v", actual)
			} else {
				s.Require().NotPanics(testFunc, "mustBzToQuarantineRecord")
				s.Assert().Equal(tc.expected, actual, "mustBzToQuarantineRecord record")
			}
		})
	}
}

func (s *TestSuite) TestQuarantineRecordGetSet() {
	s.Run("get does not exist", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr1, s.addr2)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("get multiple froms does not exist", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr1, s.addr2, s.addr3, s.addr5)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("get no froms", func() {
		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, s.addr5)
		}
		s.Assert().Panics(testFunc, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("set get one unaccepted no accepted", func() {
		toAddr := MakeTestAddr("sgouna", 0)
		uFromAddr := MakeTestAddr("sgouna", 1)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   nil,
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("foo", 123), sdk.NewInt64Coin("bar", 456)),
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual *quarantine.QuarantineRecord
		testFuncGet := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr)
		}
		var actualBackwards *quarantine.QuarantineRecord
		testFuncGetBackwards := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, uFromAddr, toAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecord") {
			s.Assert().Equal(expected, actual, "GetQuarantineRecord")
		}
		if s.Assert().NotPanics(testFuncGetBackwards, "GetQuarantineRecord wrong to/from order") {
			s.Assert().Nil(actualBackwards, "GetQuarantineRecord wrong to/from order")
		}
	})

	s.Run("set get one unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgouoa", 0)
		uFromAddr := MakeTestAddr("sgouoa", 1)
		aFromAddr := MakeTestAddr("sgouoa", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actualUA *quarantine.QuarantineRecord
		testFuncGetOrderUA := func() {
			actualUA = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr, aFromAddr)
		}
		var actualAU *quarantine.QuarantineRecord
		testFuncGetOrderAU := func() {
			actualAU = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr, uFromAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGetOrderUA, "GetQuarantineRecord order: ua") {
			s.Assert().Equal(expected, actualUA, "GetQuarantineRecord order: ua")
		}
		if s.Assert().NotPanics(testFuncGetOrderAU, "GetQuarantineRecord order: au") {
			s.Assert().Equal(expected, actualAU, "GetQuarantineRecord order: au")
		}
	})

	s.Run("set get two unaccepted no accepted", func() {
		toAddr := MakeTestAddr("sgtuna", 0)
		uFromAddr1 := MakeTestAddr("sgtuna", 1)
		uFromAddr2 := MakeTestAddr("sgtuna", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr1, uFromAddr2},
			AcceptedFromAddresses:   nil,
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual12 *quarantine.QuarantineRecord
		testFuncGet12 := func() {
			actual12 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr1, uFromAddr2)
		}
		var actual21 *quarantine.QuarantineRecord
		testFuncGet21 := func() {
			actual21 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr2, uFromAddr1)
		}
		var actualJust1 *quarantine.QuarantineRecord
		testFuncGetJust1 := func() {
			actualJust1 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr1)
		}
		var actualJust2 *quarantine.QuarantineRecord
		testFuncGetJust2 := func() {
			actualJust2 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr2)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet12, "GetQuarantineRecord order: 1 2") {
			s.Assert().Equal(expected, actual12, "GetQuarantineRecord result order: 1 2")
		}
		if s.Assert().NotPanics(testFuncGet21, "GetQuarantineRecord order: 2 1") {
			s.Assert().Equal(expected, actual21, "GetQuarantineRecord result order: 2 1")
		}
		if s.Assert().NotPanics(testFuncGetJust1, "GetQuarantineRecord just 1") {
			s.Assert().Nil(actualJust1, "GetQuarantineRecord just 1")
		}
		if s.Assert().NotPanics(testFuncGetJust2, "GetQuarantineRecord just 2") {
			s.Assert().Nil(actualJust2, "GetQuarantineRecord just 2")
		}
	})

	s.Run("set get no unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgnuoa", 0)
		aFromAddr := MakeTestAddr("sgnuoa", 1)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: nil,
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.Coins{},
			Declined:                false,
		}

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual *quarantine.QuarantineRecord
		testFuncGet := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		s.Require().NotPanics(testFuncGet, "GetQuarantineRecord")
		s.Assert().Nil(actual, "GetQuarantineRecord")
	})

	s.Run("set get no unaccepted two accepted", func() {
		toAddr := MakeTestAddr("sgnuta", 0)
		aFromAddr1 := MakeTestAddr("sgnuta", 1)
		aFromAddr2 := MakeTestAddr("sgnuta", 2)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: nil,
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr1, aFromAddr2},
			Coins:                   sdk.Coins{},
			Declined:                false,
		}

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		var actual12 *quarantine.QuarantineRecord
		testFuncGet12 := func() {
			actual12 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr1, aFromAddr2)
		}
		var actual21 *quarantine.QuarantineRecord
		testFuncGet21 := func() {
			actual21 = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, aFromAddr2, aFromAddr1)
		}

		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
		if s.Assert().NotPanics(testFuncGet12, "GetQuarantineRecord order: 1 2") {
			s.Assert().Nil(actual12, "GetQuarantineRecord order: 1 2")
		}
		if s.Assert().NotPanics(testFuncGet21, "GetQuarantineRecord order: 2 1") {
			s.Assert().Nil(actual21, "GetQuarantineRecord order: 2 1")
		}
	})

	s.Run("set get two unaccepted one accepted", func() {
		toAddr := MakeTestAddr("sgtuoa", 0)
		uFromAddr1 := MakeTestAddr("sgtuoa", 1)
		uFromAddr2 := MakeTestAddr("sgtuoa", 2)
		aFromAddr := MakeTestAddr("sgtuoa", 3)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr1, uFromAddr2},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		positiveTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1 2 a", []sdk.AccAddress{uFromAddr1, uFromAddr2, aFromAddr}},
			{"1 a 2", []sdk.AccAddress{uFromAddr1, aFromAddr, uFromAddr2}},
			{"2 1 a", []sdk.AccAddress{uFromAddr2, uFromAddr1, aFromAddr}},
			{"2 a 1", []sdk.AccAddress{uFromAddr2, aFromAddr, uFromAddr1}},
			{"a 1 2", []sdk.AccAddress{aFromAddr, uFromAddr1, uFromAddr2}},
			{"a 2 1", []sdk.AccAddress{aFromAddr, uFromAddr2, uFromAddr1}},
		}
		for _, tc := range positiveTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Equal(expected, actual, "GetQuarantineRecord")
				}
			})
		}

		negativeTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1", []sdk.AccAddress{uFromAddr1}},
			{"2", []sdk.AccAddress{uFromAddr2}},
			{"a", []sdk.AccAddress{aFromAddr}},
			{"1 1", []sdk.AccAddress{uFromAddr1, uFromAddr1}},
			{"1 2", []sdk.AccAddress{uFromAddr1, uFromAddr2}},
			{"1 a", []sdk.AccAddress{uFromAddr1, aFromAddr}},
			{"2 1", []sdk.AccAddress{uFromAddr2, uFromAddr1}},
			{"2 2", []sdk.AccAddress{uFromAddr2, uFromAddr2}},
			{"2 a", []sdk.AccAddress{uFromAddr2, aFromAddr}},
			{"a 1", []sdk.AccAddress{aFromAddr, uFromAddr1}},
			{"a 2", []sdk.AccAddress{aFromAddr, uFromAddr2}},
			{"a a", []sdk.AccAddress{aFromAddr, aFromAddr}},
			{"1 1 2", []sdk.AccAddress{uFromAddr1, uFromAddr1, uFromAddr2}},
			{"2 2 a", []sdk.AccAddress{uFromAddr2, uFromAddr2, aFromAddr}},
			{"1 2 a 1", []sdk.AccAddress{uFromAddr1, uFromAddr2, aFromAddr, uFromAddr1}},
			{"1 2 2 a", []sdk.AccAddress{uFromAddr1, uFromAddr2, uFromAddr2, aFromAddr}},
			{"a 1 2 a", []sdk.AccAddress{aFromAddr, uFromAddr1, uFromAddr2, aFromAddr}},
		}
		for _, tc := range negativeTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Nil(actual, "GetQuarantineRecord")
				}
			})
		}
	})

	s.Run("set get one unaccepted two accepted", func() {
		toAddr := MakeTestAddr("sgouta", 0)
		uFromAddr := MakeTestAddr("sgouta", 1)
		aFromAddr1 := MakeTestAddr("sgouta", 2)
		aFromAddr2 := MakeTestAddr("sgouta", 3)
		record := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr1, aFromAddr2},
			Coins:                   sdk.Coins{},
			Declined:                false,
		}
		expected := MakeCopyOfQuarantineRecord(record)

		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		positiveTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1 2 u", []sdk.AccAddress{aFromAddr1, aFromAddr2, uFromAddr}},
			{"1 u 2", []sdk.AccAddress{aFromAddr1, uFromAddr, aFromAddr2}},
			{"2 1 u", []sdk.AccAddress{aFromAddr2, aFromAddr1, uFromAddr}},
			{"2 u 1", []sdk.AccAddress{aFromAddr2, uFromAddr, aFromAddr1}},
			{"u 1 2", []sdk.AccAddress{uFromAddr, aFromAddr1, aFromAddr2}},
			{"u 2 1", []sdk.AccAddress{uFromAddr, aFromAddr2, aFromAddr1}},
		}
		for _, tc := range positiveTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Equal(expected, actual, "GetQuarantineRecord")
				}
			})
		}

		negativeTests := []struct {
			name      string
			fromAddrs []sdk.AccAddress
		}{
			{"1", []sdk.AccAddress{aFromAddr1}},
			{"2", []sdk.AccAddress{aFromAddr2}},
			{"u", []sdk.AccAddress{uFromAddr}},
			{"1 1", []sdk.AccAddress{aFromAddr1, aFromAddr1}},
			{"1 2", []sdk.AccAddress{aFromAddr1, aFromAddr2}},
			{"1 u", []sdk.AccAddress{aFromAddr1, uFromAddr}},
			{"2 1", []sdk.AccAddress{aFromAddr2, aFromAddr1}},
			{"2 2", []sdk.AccAddress{aFromAddr2, aFromAddr2}},
			{"2 u", []sdk.AccAddress{aFromAddr2, uFromAddr}},
			{"u 1", []sdk.AccAddress{uFromAddr, aFromAddr1}},
			{"u 2", []sdk.AccAddress{uFromAddr, aFromAddr2}},
			{"u u", []sdk.AccAddress{uFromAddr, uFromAddr}},
			{"1 1 2", []sdk.AccAddress{aFromAddr1, aFromAddr1, aFromAddr2}},
			{"2 2 u", []sdk.AccAddress{aFromAddr2, aFromAddr2, uFromAddr}},
			{"1 2 u 1", []sdk.AccAddress{aFromAddr1, aFromAddr2, uFromAddr, aFromAddr1}},
			{"1 2 2 u", []sdk.AccAddress{aFromAddr1, aFromAddr2, aFromAddr2, uFromAddr}},
			{"u 1 2 u", []sdk.AccAddress{uFromAddr, aFromAddr1, uFromAddr, uFromAddr}},
		}
		for _, tc := range negativeTests {
			s.Run("GetQuarantineRecord order "+tc.name, func() {
				var actual *quarantine.QuarantineRecord
				testFunc := func() {
					actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, tc.fromAddrs...)
				}
				if s.Assert().NotPanics(testFunc, "GetQuarantineRecord") {
					s.Assert().Nil(actual, "GetQuarantineRecord")
				}
			})
		}
	})
}

func (s *TestSuite) TestGetQuarantineRecords() {
	addr0 := MakeTestAddr("gqr", 0)
	addr1 := MakeTestAddr("gqr", 1)
	addr2 := MakeTestAddr("gqr", 2)
	addr3 := MakeTestAddr("gqr", 3)

	mustCoins := func(amt string) sdk.Coins {
		coins, err := sdk.ParseCoinsNormalized(amt)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", amt)
		return coins
	}

	recordA := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: []sdk.AccAddress{addr1},
		Coins:                   mustCoins("1acoin"),
		Declined:                true,
	}
	recordB := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: []sdk.AccAddress{addr1},
		AcceptedFromAddresses:   []sdk.AccAddress{addr2},
		Coins:                   mustCoins("10bcoin"),
	}
	recordC := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: []sdk.AccAddress{addr2},
		Coins:                   mustCoins("100ccoin"),
	}
	recordD := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: []sdk.AccAddress{addr1},
		AcceptedFromAddresses:   []sdk.AccAddress{addr0, addr2},
		Coins:                   mustCoins("1000dcoin"),
	}
	recordE := &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: []sdk.AccAddress{addr3},
		Coins:                   mustCoins("100000ecoin"),
	}

	testFunc := func(toAddr sdk.AccAddress, record *quarantine.QuarantineRecord) func() {
		return func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
		}
	}

	s.Require().NotPanics(testFunc(addr0, recordA), "SetQuarantineRecord recordA")
	s.Require().NotPanics(testFunc(addr0, recordB), "SetQuarantineRecord recordB")
	s.Require().NotPanics(testFunc(addr0, recordC), "SetQuarantineRecord recordC")
	s.Require().NotPanics(testFunc(addr0, recordD), "SetQuarantineRecord recordD")
	s.Require().NotPanics(testFunc(addr0, recordE), "SetQuarantineRecord recordE")

	// Setup:
	// 0 <- 1:  1acoin declined
	// 0 <- 1 2: 10bcoin
	// 0 <- 2: 100ccoin
	// 0 <- 0 1 2: 1000dcoin
	// 0 <- 3: 10000ecoin

	addrs := func(addz ...sdk.AccAddress) []sdk.AccAddress {
		return addz
	}

	qrs := func(qrz ...*quarantine.QuarantineRecord) []*quarantine.QuarantineRecord {
		return qrz
	}

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		expected  []*quarantine.QuarantineRecord
	}{
		{
			name:      "to 0 from none",
			toAddr:    addr0,
			fromAddrs: addrs(),
			expected:  nil,
		},
		{
			name:      "to 1 from none",
			toAddr:    addr0,
			fromAddrs: addrs(),
			expected:  nil,
		},
		{
			name:      "to 2 from none",
			toAddr:    addr0,
			fromAddrs: addrs(),
			expected:  nil,
		},
		{
			name:      "to 3 from none",
			toAddr:    addr0,
			fromAddrs: addrs(),
			expected:  nil,
		},
		{
			name:      "to 1 from 0",
			toAddr:    addr1,
			fromAddrs: addrs(addr0),
			expected:  nil,
		},
		{
			name:      "to 1 from 1",
			toAddr:    addr1,
			fromAddrs: addrs(addr1),
			expected:  nil,
		},
		{
			name:      "to 1 from 2",
			toAddr:    addr1,
			fromAddrs: addrs(addr2),
			expected:  nil,
		},
		{
			name:      "to 1 from 3",
			toAddr:    addr1,
			fromAddrs: addrs(addr3),
			expected:  nil,
		},
		{
			name:      "to 2 from 0",
			toAddr:    addr2,
			fromAddrs: addrs(addr0),
			expected:  nil,
		},
		{
			name:      "to 2 from 1",
			toAddr:    addr2,
			fromAddrs: addrs(addr1),
			expected:  nil,
		},
		{
			name:      "to 2 from 2",
			toAddr:    addr2,
			fromAddrs: addrs(addr2),
			expected:  nil,
		},
		{
			name:      "to 2 from 3",
			toAddr:    addr2,
			fromAddrs: addrs(addr3),
			expected:  nil,
		},
		{
			name:      "to 3 from 0",
			toAddr:    addr3,
			fromAddrs: addrs(addr0),
			expected:  nil,
		},
		{
			name:      "to 3 from 1",
			toAddr:    addr3,
			fromAddrs: addrs(addr1),
			expected:  nil,
		},
		{
			name:      "to 3 from 2",
			toAddr:    addr3,
			fromAddrs: addrs(addr2),
			expected:  nil,
		},
		{
			name:      "to 3 from 3",
			toAddr:    addr3,
			fromAddrs: addrs(addr3),
			expected:  nil,
		},
		{
			name:      "to 3 from 0 1 2 3",
			toAddr:    addr3,
			fromAddrs: addrs(addr0, addr1, addr2, addr3),
			expected:  nil,
		},
		{
			name:      "to 0 from 0 finds 1: d",
			toAddr:    addr0,
			fromAddrs: addrs(addr0),
			expected:  qrs(recordD),
		},
		{
			name:      "to 0 from 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: addrs(addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 3 finds 1: e",
			toAddr:    addr0,
			fromAddrs: addrs(addr3),
			expected:  qrs(recordE),
		},
		{
			name:      "to 0 from 0 0 finds 1: d",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr0),
			expected:  qrs(recordD),
		},
		{
			name:      "to 0 from 0 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 0 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 0 3 finds 2: de",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr3),
			expected:  qrs(recordD, recordE),
		},
		{
			name:      "to 0 from 1 0 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr0),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 1 1 finds 3: abd",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr1),
			expected:  qrs(recordA, recordB, recordD),
		},
		{
			name:      "to 0 from 1 2 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 1 3 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr3),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 2 0 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr2, addr0),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 1 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr2, addr1),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 2 finds 3: bcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr2, addr2),
			expected:  qrs(recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 2 3 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr2, addr3),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 3 0 finds 2: de",
			toAddr:    addr0,
			fromAddrs: addrs(addr3, addr0),
			expected:  qrs(recordD, recordE),
		},
		{
			name:      "to 0 from 3 1 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: addrs(addr3, addr1),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 3 2 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr3, addr2),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 3 3 finds 1: e",
			toAddr:    addr0,
			fromAddrs: addrs(addr3, addr3),
			expected:  qrs(recordE),
		},
		{
			name:      "to 0 from 0 1 2 finds 4: abcd",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD),
		},
		{
			name:      "to 0 from 0 1 3 finds 4: abde",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr1, addr3),
			expected:  qrs(recordA, recordB, recordD, recordE),
		},
		{
			name:      "to 0 from 0 2 3 finds 4: bcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr2, addr3),
			expected:  qrs(recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 2 3 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr2, addr3),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 0 1 2 3 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr0, addr1, addr2, addr3),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 3 0 2 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr3, addr0, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
		{
			name:      "to 0 from 1 3 1 2 finds 5: abcde",
			toAddr:    addr0,
			fromAddrs: addrs(addr1, addr3, addr1, addr2),
			expected:  qrs(recordA, recordB, recordC, recordD, recordE),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual := s.keeper.GetQuarantineRecords(s.sdkCtx, tc.toAddr, tc.fromAddrs...)
			s.Assert().ElementsMatch(tc.expected, actual, "GetQuarantineRecords A = expected vs B = actual")
		})
	}
}

func (s *TestSuite) TestAddQuarantinedCoins() {
	makeEvents := func(toAddr sdk.AccAddress, coins sdk.Coins) (sdk.Events, error) {
		events, err := sdk.TypedEventToEvent(&quarantine.EventFundsQuarantined{
			ToAddress: toAddr.String(),
			Coins:     coins,
		})
		return sdk.Events{events}, err
	}

	addAndCheck := func(t *testing.T, amt string, toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) bool {
		t.Helper()
		coins, err := sdk.ParseCoinsNormalized(amt)
		if !assert.NoError(t, err, "ParseCoinsNormalized(%s)", amt) {
			return false
		}
		var expEvents sdk.Events
		expEvents, err = makeEvents(toAddr, coins)
		if !assert.NoError(t, err, "creating expected events") {
			return false
		}
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		testFuncAdd := func() {
			err = s.keeper.AddQuarantinedCoins(ctx, coins, toAddr, fromAddrs...)
		}
		if !assert.NotPanics(t, testFuncAdd, "AddQuarantinedCoins") {
			return false
		}
		if !assert.NoError(t, err, "AddQuarantinedCoins") {
			return false
		}
		actEvents := ctx.EventManager().Events()
		return assert.Equal(t, expEvents, actEvents)
	}

	s.Run("no record yet new record created", func() {
		toAddr := MakeTestAddr("nrynrc", 0)
		fromAddr := MakeTestAddr("nrynrc", 1)
		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("bananas", 99)),
			Declined:                false,
		}

		addAndCheck(s.T(), "99bananas", toAddr, fromAddr)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	s.Run("no record yet multiple froms new record created", func() {
		toAddr := MakeTestAddr("nrymfnrc", 0)
		fromAddr1 := MakeTestAddr("nrymfnrc", 1)
		fromAddr2 := MakeTestAddr("nrymfnrc", 2)
		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr1, fromAddr2},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("bananas", 99)),
			Declined:                false,
		}

		addAndCheck(s.T(), "99bananas", toAddr, fromAddr1, fromAddr2)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddr1, fromAddr2)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	s.Run("record exists is updated same denom", func() {
		toAddr := MakeTestAddr("reiusd", 0)
		fromAddr := MakeTestAddr("reiusd", 1)
		orig := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pants", 11)),
			Declined:                false,
		}
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, orig)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pants", 211)),
			Declined:                false,
		}

		addAndCheck(s.T(), "200pants", toAddr, fromAddr)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	s.Run("record exists is updated different denom", func() {
		toAddr := MakeTestAddr("reiudd", 0)
		fromAddr := MakeTestAddr("reiudd", 1)
		orig := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pants", 11)),
			Declined:                false,
		}
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, orig)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{fromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pants", 11), sdk.NewInt64Coin("hatcoins", 250)),
			Declined:                false,
		}

		addAndCheck(s.T(), "250hatcoins", toAddr, fromAddr)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddr)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	s.Run("record exists is updated multiple froms", func() {
		toAddr := MakeTestAddr("reiumf", 0)
		uFromAddr := MakeTestAddr("reiumf", 1)
		aFromAddr := MakeTestAddr("reiumf", 2)
		orig := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 53)),
			Declined:                false,
		}
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, orig)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 9053)),
			Declined:                false,
		}

		addAndCheck(s.T(), "9000pcoin", toAddr, uFromAddr, aFromAddr)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr, aFromAddr)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	s.Run("record exists is updated multiple froms other order", func() {
		toAddr := MakeTestAddr("reiumfo", 0)
		uFromAddr := MakeTestAddr("reiumfo", 1)
		aFromAddr := MakeTestAddr("reiumfo", 2)
		orig := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 35)),
			Declined:                false,
		}
		testFuncSet := func() {
			s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, orig)
		}
		s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")

		expected := &quarantine.QuarantineRecord{
			UnacceptedFromAddresses: []sdk.AccAddress{uFromAddr},
			AcceptedFromAddresses:   []sdk.AccAddress{aFromAddr},
			Coins:                   sdk.NewCoins(sdk.NewInt64Coin("pcoin", 935)),
			Declined:                false,
		}

		addAndCheck(s.T(), "900pcoin", toAddr, aFromAddr, uFromAddr)

		var actual *quarantine.QuarantineRecord
		testFunc := func() {
			actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, uFromAddr, aFromAddr)
		}
		s.Require().NotPanics(testFunc, "GetQuarantineRecord")
		s.Assert().Equal(expected, actual, "resulting quarantine record")
	})

	// one from, is auto-accept, nothing stored.
	// two froms, one is auto-accept, is marked as such.
	// two froms, both auto-accept, nothing stored.
	// existing record not declined, from not auto-decline, result is not declined.
	// existing record not declined, from is auto-decline, result is declined.
	// existing record declined, from is auto-declined, result is declined.
	// existing record declined, form is not auto-declined, result is not declined.
	// existing record not declined, 2 froms, neither are auto-decline, result is not declined.
	// existing record not declined, 2 froms, first is auto-decline, result is declined.
	// existing record not declined, 2 froms, second is auto-decline, result is declined.
	// existing record not declined, 2 froms, both are auto-decline, result is declined.
	// existing record declined, 2 froms, neither are auto-decline, result is not declined
	// existing record declined, 2 froms, first is auto-decline, result is declined.
	// existing record declined, 2 froms, second is auto-decline, result is declined.
	// existing record declined, 2 froms, both are auto-decline, result is declined.
}

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
