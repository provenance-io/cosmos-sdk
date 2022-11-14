package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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

// czt is a way to create coins that requires fewer characters than sdk.NewCoins(sdk.NewInt64Coin("foo", 5))
func czt(t *testing.T, coins string) sdk.Coins {
	rv, err := sdk.ParseCoinsNormalized(coins)
	require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
	return rv
}

// updateQR updates the AccAddresses using the provided addrs.
// Any AccAddress that is 1 byte long and can be an index in addrs,
// is replaced by the addrs entry using that byte as the index.
// E.g. if UnacceptedFromAddresses is []sdk.AccAddress{{1}}, then it will be replaced with addrs[1].
func updateQR(addrs []sdk.AccAddress, record *quarantine.QuarantineRecord) {
	if record != nil {
		for i, addr := range record.UnacceptedFromAddresses {
			if len(addr) == 1 && int(addr[0]) < len(addrs) {
				record.UnacceptedFromAddresses[i] = addrs[addr[0]]
			}
		}
		for i, addr := range record.AcceptedFromAddresses {
			if len(addr) == 1 && int(addr[0]) < len(addrs) {
				record.AcceptedFromAddresses[i] = addrs[addr[0]]
			}
		}
	}
}

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
	// cz an even shorter way of creating coins since all creating should get the same *testing.T here.
	cz := func(coins string) sdk.Coins {
		return czt(s.T(), coins)
	}

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
				Coins:                   cz("9000bar,888foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   cz("9000bar,888foo"),
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
				Coins:                   cz("9001bar,889foo"),
				Declined:                true,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr1},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2},
				Coins:                   cz("9001bar,889foo"),
				Declined:                true,
			},
		},
		{
			name: "no unaccepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{},
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2, s.addr1, s.addr3},
				Coins:                   cz("9002bar,890foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: nil,
				AcceptedFromAddresses:   []sdk.AccAddress{s.addr2, s.addr1, s.addr3},
				Coins:                   cz("9002bar,890foo"),
				Declined:                false,
			},
		},
		{
			name: "no accepted",
			bz: cdc.MustMarshal(&quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr4, s.addr2, s.addr5},
				AcceptedFromAddresses:   []sdk.AccAddress{},
				Coins:                   cz("9003bar,891foo"),
				Declined:                false,
			}),
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{s.addr4, s.addr2, s.addr5},
				AcceptedFromAddresses:   nil,
				Coins:                   cz("9003bar,891foo"),
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
			Coins:                   czt(s.T(), "456bar,1233foo"),
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
	// cz an even shorter way of creating coins since all creating should get the same *testing.T here.
	cz := func(coins string) sdk.Coins {
		return czt(s.T(), coins)
	}
	// Getting a little tricky here because I want different addresses for each test.
	// The addrBase is used to generate addrCount addresses.
	// Then, the autoAccept, autoDecline, toAddr and fromAddrs are address indexes to use.
	// The tricky part is that both the existing and expected Quarantine Records will have their
	// AccAddress slices updated before doing anything. For any AccAddress in them that's 1 byte long, and that byte
	// is less than addrCount, it's used as an index and the entry is updated to be that address.
	tests := []struct {
		name        string
		addrBase    string
		addrCount   uint8
		existing    *quarantine.QuarantineRecord
		autoAccept  []int
		autoDecline []int
		coins       sdk.Coins
		toAddr      int
		fromAddrs   []int
		expected    *quarantine.QuarantineRecord
	}{
		{
			name:      "new record is created",
			addrBase:  "nr",
			addrCount: 2,
			coins:     cz("99bananas"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("99bananas"),
			},
		},
		{
			name:      "new record 2 froms is created",
			addrBase:  "nr2f",
			addrCount: 3,
			coins:     cz("88crazy"),
			toAddr:    0,
			fromAddrs: []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("88crazy"),
			},
		},
		{
			name:      "existing record same denom is updated",
			addrBase:  "ersd",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("11pants"),
			},
			coins:     cz("200pants"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("211pants"),
			},
		},
		{
			name:      "existing record new denom is updated",
			addrBase:  "ernd",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("102tower"),
			},
			coins:     cz("5pit"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("5pit,102tower"),
			},
		},
		{
			name:      "existing record 2 froms is updated",
			addrBase:  "er2f",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   cz("53pcoin"),
			},
			coins:     cz("9000pcoin"),
			toAddr:    0,
			fromAddrs: []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   cz("9053pcoin"),
			},
		},
		{
			name:      "existing record 2 froms other order is updated",
			addrBase:  "er2foo",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   cz("35pcoin"),
			},
			coins:     cz("800pcoin"),
			toAddr:    0,
			fromAddrs: []int{2, 1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   cz("835pcoin"),
			},
		},
		{
			name:      "existing record unaccepted now auto-accept is still unaccepted",
			addrBase:  "eruna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("543interstellar"),
			},
			autoAccept: []int{1},
			coins:      cz("5012interstellar"),
			toAddr:     0,
			fromAddrs:  []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("5555interstellar"), // One more time!
			},
		},
		{
			name:       "new record from is auto-accept nothing stored",
			addrBase:   "nrfa",
			addrCount:  2,
			autoAccept: []int{1},
			coins:      cz("76trombones"),
			toAddr:     0,
			fromAddrs:  []int{1},
			expected:   nil,
		},
		{
			name:       "new record two froms first is auto-accept is marked as such",
			addrBase:   "nr2fa",
			addrCount:  3,
			autoAccept: []int{1},
			coins:      cz("52pinata"),
			toAddr:     0,
			fromAddrs:  []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{2}},
				AcceptedFromAddresses:   []sdk.AccAddress{{1}},
				Coins:                   cz("52pinata"),
			},
		},
		{
			name:       "new record two froms second is auto-accept is marked as such",
			addrBase:   "nr2sa",
			addrCount:  3,
			autoAccept: []int{2},
			coins:      cz("3fiddy"), // Loch Ness Monster, is that you?
			toAddr:     0,
			fromAddrs:  []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				AcceptedFromAddresses:   []sdk.AccAddress{{2}},
				Coins:                   cz("3fiddy"),
			},
		},
		{
			name:       "new record two froms both auto-accept nothing stored",
			addrBase:   "nr2ba",
			addrCount:  3,
			autoAccept: []int{1, 2},
			coins:      cz("4moo"),
			toAddr:     0,
			fromAddrs:  []int{1, 2},
			expected:   nil,
		},
		{
			name:      "existing record not declined not auto-decline result is not declined",
			addrBase:  "erndna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("8nodeca"),
				Declined:                false,
			},
			coins:     cz("50nodeca"),
			toAddr:    0,
			fromAddrs: []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("58nodeca"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined is auto-decline result is declined",
			addrBase:  "erndad",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("20deca"),
				Declined:                false,
			},
			autoDecline: []int{1},
			coins:       cz("406deca"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("426deca"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined is auto-declined result is declined",
			addrBase:  "erdad",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("3000yarp"),
				Declined:                true,
			},
			autoDecline: []int{1},
			coins:       cz("3yarp"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("3003yarp"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined not auto-declined result is not declined",
			addrBase:  "erdna",
			addrCount: 2,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("14dalmatian"),
				Declined:                true,
			},
			autoDecline: nil,
			coins:       cz("87dalmatian"),
			toAddr:      0,
			fromAddrs:   []int{1},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}},
				Coins:                   cz("101dalmatian"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined 2 froms neither are auto-decline result is not declined",
			addrBase:  "ernd2fna",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("3bill"),
				Declined:                false,
			},
			autoDecline: nil,
			coins:       cz("4bill"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("7bill"),
				Declined:                false,
			},
		},
		{
			name:      "existing record not declined 2 froms first is auto-decline result is declined",
			addrBase:  "ernd2ffa",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("20123zela"),
				Declined:                false,
			},
			autoDecline: []int{1},
			coins:       cz("5000000000zela"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("5000020123zela"),
				Declined:                true,
			},
		},
		{
			name:      "existing record not declined 2 froms second is auto-decline result is declined",
			addrBase:  "ernd2fsd",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("456789vids"),
				Declined:                false,
			},
			autoDecline: []int{2},
			coins:       cz("123000000vids"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("123456789vids"),
				Declined:                true,
			},
		},
		{
			name:      "existing record not declined 2 froms both are auto-decline result is declined",
			addrBase:  "ernd2fba",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("5green"),
				Declined:                false,
			},
			autoDecline: []int{1, 2},
			coins:       cz("333333333333333green"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("333333333333338green"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms neither are auto-decline result is not declined",
			addrBase:  "erd2fna",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("4frank"),
				Declined:                true,
			},
			autoDecline: nil,
			coins:       cz("3frank"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("7frank"),
				Declined:                false,
			},
		},
		{
			name:      "existing record declined 2 froms first is auto-decline result is declined",
			addrBase:  "erd2ffa",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("10zulu"),
				Declined:                true,
			},
			autoDecline: []int{1},
			coins:       cz("11zulu"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("21zulu"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms second is auto-decline result is declined",
			addrBase:  "erd2fsd",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("11stars"),
				Declined:                true,
			},
			autoDecline: []int{2},
			coins:       cz("99stars"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("110stars"),
				Declined:                true,
			},
		},
		{
			name:      "existing record declined 2 froms both are auto-decline result is declined",
			addrBase:  "erd2fba",
			addrCount: 3,
			existing: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("44blue"),
				Declined:                true,
			},
			autoDecline: []int{1, 2},
			coins:       cz("360blue"),
			toAddr:      0,
			fromAddrs:   []int{1, 2},
			expected: &quarantine.QuarantineRecord{
				UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
				Coins:                   cz("404blue"),
				Declined:                true,
			},
		},
	}

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}
			toAddr := addrs[tc.toAddr]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}
			autoAccept := make([]sdk.AccAddress, len(tc.autoAccept))
			for i, ai := range tc.autoAccept {
				autoAccept[i] = addrs[ai]
			}
			autoDecline := make([]sdk.AccAddress, len(tc.autoDecline))
			for i, ai := range tc.autoDecline {
				autoDecline[i] = addrs[ai]
			}
			updateQR(addrs, tc.existing)
			updateQR(addrs, tc.expected)

			// Set the existing value
			if tc.existing != nil {
				testFuncSet := func() {
					s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, tc.existing)
				}
				s.Require().NotPanics(testFuncSet, "SetQuarantineRecord")
			}

			// Set up auto-accept and auto-decline
			testFuncAuto := func(fromAddr sdk.AccAddress, response quarantine.AutoResponse) func() {
				return func() {
					s.keeper.SetAutoResponse(s.sdkCtx, toAddr, fromAddr, response)
				}
			}
			for i, fromAddr := range autoAccept {
				s.Require().NotPanics(testFuncAuto(fromAddr, quarantine.AUTO_RESPONSE_ACCEPT), "SetAutoResponse %d accept", i+1)
			}
			for i, fromAddr := range autoDecline {
				s.Require().NotPanics(testFuncAuto(fromAddr, quarantine.AUTO_RESPONSE_DECLINE), "SetAutoResponse %d decline", i+1)
			}

			// Create events expected to be emitted by AddQuarantinedCoins.
			event, err := sdk.TypedEventToEvent(&quarantine.EventFundsQuarantined{
				ToAddress: toAddr.String(),
				Coins:     tc.coins,
			})
			s.Require().NoError(err, "TypedEventToEvent EventFundsQuarantined")
			expectedEvents := sdk.Events{event}

			// Get a context with a fresh event manager and call AddQuarantinedCoins.
			// Make sure it doesn't panic and make sure it doesn't return an error.
			// Note: As of writing, the only error it could return is from emitting the events,
			// and who knows how to actually trigger/test that.
			ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
			testFuncAdd := func() {
				err = s.keeper.AddQuarantinedCoins(ctx, tc.coins, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncAdd, "AddQuarantinedCoins")
			s.Require().NoError(err, "AddQuarantinedCoins")
			actualEvents := ctx.EventManager().Events()
			s.Assert().Equal(expectedEvents, actualEvents)

			// Now look up the record and make sure it's as expected.
			var actual *quarantine.QuarantineRecord
			testFuncGet := func() {
				actual = s.keeper.GetQuarantineRecord(s.sdkCtx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncGet, "GetQuarantineRecord")
			s.Assert().Equal(tc.expected, actual, "resulting quarantine record")
		})
	}
}

func (s *TestSuite) TestAcceptQuarantinedFunds() {
	// cz an even shorter way of creating coins since all creating should get the same *testing.T here.
	cz := func(coins string) sdk.Coins {
		return czt(s.T(), coins)
	}

	// makeEvent creates a funds-released event.
	makeEvent := func(t *testing.T, addr sdk.AccAddress, amt sdk.Coins) sdk.Event {
		event, err := sdk.TypedEventToEvent(&quarantine.EventFundsReleased{
			ToAddress: addr.String(),
			Coins:     amt,
		})
		require.NoError(t, err, "TypedEventToEvent EventFundsReleased")
		return event
	}

	// An event maker knows the coins, and takes in the address to output an
	// event with the (presently unknown) ToAddress and the (known) coins.
	type eventMaker func(t *testing.T, addr sdk.AccAddress) sdk.Event

	// makes the event maker functions, one for each string provided.
	makeEventMakers := func(coins ...string) []eventMaker {
		rv := make([]eventMaker, len(coins))
		for i, amtStr := range coins {
			// doing this now so that an invalid coin string fails the test before it gets started.
			// Really, I didn't want to have to update cz to also take in a *testing.T.
			amt := cz(amtStr)
			rv[i] = func(t *testing.T, addr sdk.AccAddress) sdk.Event {
				return makeEvent(t, addr, amt)
			}
		}
		return rv
	}

	// Getting a little tricky here because I want different addresses for each test.
	// The addrBase is used to generate addrCount addresses.
	// Then, addrs[0] becomes the toAddr. The fromAddrs are indexes of the addrs to use.
	// The tricky part is that the existing and expected Quarantine Records will have their
	// AccAddresses updated before doing anything. For any AccAddress in them that's 1 byte long, and that byte
	// is less than addrCount, it's used as an index and the entry is updated to be that address.
	// Also, the provided []eventMaker is used to create all expected events receiving the toAddr.
	tests := []struct {
		name            string
		addrBase        string
		addrCount       uint8
		records         []*quarantine.QuarantineRecord
		autoDecline     []int
		fromAddrs       []int
		expectedRecords []*quarantine.QuarantineRecord
		expectedSent    []sdk.Coins
		expectedEvents  []eventMaker
	}{
		{
			name:            "one from zero records",
			addrBase:        "ofzr",
			addrCount:       2,
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    nil,
			expectedEvents:  nil,
		},
		{
			name:      "one from one record fully",
			addrBase:  "oforf",
			addrCount: 2,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("17lemon"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("17lemon")},
			expectedEvents:  makeEventMakers("17lemon"),
		},
		{
			name:      "one from one record finally fully",
			addrBase:  "foforf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}, {3}},
					Coins:                   cz("8878pillow"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("8878pillow")},
			expectedEvents:  makeEventMakers("8878pillow"),
		},
		{
			name:      "one from one record fully previously declined",
			addrBase:  "oforfpd",
			addrCount: 2,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("5rings,4birds,3hens"),
					Declined:                true,
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("5rings,4birds,3hens")},
			expectedEvents:  makeEventMakers("5rings,4birds,3hens"),
		},
		{
			name:      "one from one record not fully",
			addrBase:  "ofornf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("1snow"),
					Declined:                false,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("1snow"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record not fully previously declined",
			addrBase:  "ofornfpd",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("55orchid"),
					Declined:                true,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("55orchid"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record remaining unaccepted is auto-decline",
			addrBase:  "oforruad",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("99redballoons"),
					Declined:                true,
				},
			},
			autoDecline: []int{2},
			fromAddrs:   []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("99redballoons"),
					Declined:                true,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from one record accepted was auto-decline",
			addrBase:  "oforawad",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("7777frog"),
					Declined:                true,
				},
			},
			autoDecline: []int{1},
			fromAddrs:   []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("7777frog"),
					Declined:                false,
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from two records neither fully",
			addrBase:  "oftrnf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("20533lamp"),
					Declined:                true,
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   cz("45sun"),
					Declined:                true,
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("20533lamp"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("45sun"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "one from two records first fully",
			addrBase:  "oftrff",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 0264500F71512C3B111D2D2EAA7322F018DA16B13CBB5D516BD4B51C4F1A94EC
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   cz("43bulb"),
				},
				// key suffix = 47F604CA662719863E40CF215D4DE088C22B7FF217236D887A99AF63A8F124E9
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   cz("5005shade"),
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("5005shade"),
				},
			},
			expectedSent:   []sdk.Coins{cz("43bulb")},
			expectedEvents: makeEventMakers("43bulb"),
		},
		{
			name:      "one from two records second fully",
			addrBase:  "ofttrsf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = EFC545E02C1785EEAAE9004385C6106E75AC42E8096556376097037A0C122E41
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   cz("346awning"),
				},
				// key suffix = F898B0EAF64B4D67BC2C285E541D381FA422D85B05C69D697C099B1968003955
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}},
					Coins:                   cz("9444sprout"),
				},
			},
			fromAddrs: []int{1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("346awning"),
				},
			},
			expectedSent:   []sdk.Coins{cz("9444sprout")},
			expectedEvents: makeEventMakers("9444sprout"),
		},
		{
			name:      "one from two records both fully",
			addrBase:  "oftrbf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("4312stand"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					AcceptedFromAddresses:   []sdk.AccAddress{{2}, {3}},
					Coins:                   cz("9867sit"),
				},
			},
			fromAddrs:       []int{1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("4312stand"), cz("9867sit")},
			expectedEvents:  makeEventMakers("4312stand", "9867sit"),
		},
		{
			name:            "two froms zero records",
			addrBase:        "tfzr",
			addrCount:       3,
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    nil,
			expectedEvents:  nil,
		},
		{
			name:      "two froms one record fully",
			addrBase:  "tforf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("838hibiscus"),
				},
			},
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("838hibiscus")},
			expectedEvents:  makeEventMakers("838hibiscus"),
		},
		{
			name:      "two froms other order one record fully",
			addrBase:  "tfooorf",
			addrCount: 3,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("10downing"),
				},
			},
			fromAddrs:       []int{2, 1},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("10downing")},
			expectedEvents:  makeEventMakers("10downing"),
		},
		{
			name:      "two froms one record not fully",
			addrBase:  "tfornf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   cz("1060waddison"),
				},
			},
			fromAddrs: []int{1, 2},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("1060waddison"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms other order one record not fully",
			addrBase:  "tfooornf",
			addrCount: 4,
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   cz("1060waddison"),
				},
			},
			fromAddrs: []int{2, 1},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("1060waddison"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms two records neither fully",
			addrBase:  "tftrnf",
			addrCount: 5,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 70705D4547681D550CF0D2A5B0996B6C2B42E181FF3F84A71CF6DAD8527C8C9C
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {4}},
					Coins:                   cz("12drummers"),
				},
				// key suffix = 83A580037E196C7BB4B36FDB5531BA715DF24F86681A61FE7D72D77BE2ABA4E8
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}, {3}},
					Coins:                   cz("11pipers"),
				},
			},
			fromAddrs: []int{1, 2},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{4}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("12drummers"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{3}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("11pipers"),
				},
			},
			expectedSent:   nil,
			expectedEvents: nil,
		},
		{
			name:      "two froms two records first fully",
			addrBase:  "tftrff",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 72536EA1F5EB0C1FF2897309892EF28553E7A6C2508AB1751D363B8C3A31A56F
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   cz("8maids,7swans"),
				},
				// key suffix = BDA18A04E7AC80DDA290C262CBEF7C2928B95F9DBFE8F392BA82EC0186DBA0CC
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("10lords,9ladies"),
				},
			},
			fromAddrs: []int{1, 3},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("10lords,9ladies"),
				},
			},
			expectedSent:   []sdk.Coins{cz("8maids,7swans")},
			expectedEvents: makeEventMakers("8maids,7swans"),
		},
		{
			name:      "two froms two records second fully",
			addrBase:  "tftrsf",
			addrCount: 4,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				// key suffix = 00E641E0BF6DF9F97E61B94BBBA58B78F74198BB72681C9A24C12D2BF1DDC371
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {2}},
					Coins:                   cz("6geese"),
				},
				// key suffix = D052411A78E6208D482F600692C7382C814C35FB75B49430E5CF895B4FE5EEFF
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}, {3}},
					Coins:                   cz("2doves,1peartree"),
				},
			},
			fromAddrs: []int{1, 3},
			expectedRecords: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					AcceptedFromAddresses:   []sdk.AccAddress{{1}},
					Coins:                   cz("6geese"),
				},
			},
			expectedSent:   []sdk.Coins{cz("2doves,1peartree")},
			expectedEvents: makeEventMakers("2doves,1peartree"),
		},
		{
			name:      "two froms two records both fully",
			addrBase:  "tftrbf",
			addrCount: 3,
			// Note: This assumes AcceptQuarantinedFunds loops through the records ordered by key.
			//       The ordering defined here should match that to make test maintenance easier.
			records: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("3amigos"),
				},
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{2}},
					Coins:                   cz("8amigos"),
				},
			},
			fromAddrs:       []int{1, 2},
			expectedRecords: nil,
			expectedSent:    []sdk.Coins{cz("3amigos"), cz("8amigos")},
			expectedEvents:  makeEventMakers("3amigos", "8amigos"),
		},
	}

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "", "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}

			toAddr := addrs[0]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}

			autoDecline := make([]sdk.AccAddress, len(tc.autoDecline))
			for i, addr := range tc.autoDecline {
				autoDecline[i] = addrs[addr]
			}

			for _, record := range tc.records {
				updateQR(addrs, record)
			}

			for _, record := range tc.expectedRecords {
				updateQR(addrs, record)
			}

			var expectedSends []*SentCoins
			if len(tc.expectedSent) > 0 {
				expectedSends = make([]*SentCoins, len(tc.expectedSent))
				for i, sent := range tc.expectedSent {
					expectedSends[i] = &SentCoins{
						FromAddr: s.keeper.GetFundsHolder(),
						ToAddr:   toAddr,
						Amt:      sent,
					}
				}
			}

			expectedEvents := make(sdk.Events, len(tc.expectedEvents))
			for i, ev := range tc.expectedEvents {
				expectedEvents[i] = ev(s.T(), toAddr)
			}

			// Now that we have all the expected stuff defined, let's get things set up.

			// mock the bank keeper and use that, so we don't have to fund stuff,
			// and we get a record of the sends made.
			bKeeper := NewMockBankKeeper() // bzzzzzzzzzz
			qKeeper := s.keeper.WithBankKeeper(bKeeper)

			// Set the existing records
			for i, existing := range tc.records {
				if existing != nil {
					testFuncSet := func() {
						qKeeper.SetQuarantineRecord(s.sdkCtx, toAddr, existing)
					}
					s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
					recordKey := quarantine.CreateRecordKey(toAddr, existing.GetAllFromAddrs()...)
					_, suffix := quarantine.ParseRecordIndexKey(recordKey)
					s.T().Logf("existing[%d] suffix: %v", i, suffix)
				}
			}

			// Set existing auto-declines
			for i, addr := range autoDecline {
				testFuncAuto := func() {
					qKeeper.SetAutoResponse(s.sdkCtx, toAddr, addr, quarantine.AUTO_RESPONSE_DECLINE)
				}
				s.Require().NotPanics(testFuncAuto, "SetAutoResponse[%d]", i)
			}

			// Setup done. Let's do this.
			var err error
			ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
			testFuncAccept := func() {
				err = qKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncAccept, "AcceptQuarantinedFunds")
			s.Require().NoError(err, "AcceptQuarantinedFunds")

			// And check the expected.
			var actualRecords []*quarantine.QuarantineRecord
			testFuncGet := func() {
				actualRecords = qKeeper.GetQuarantineRecords(s.sdkCtx, toAddr, fromAddrs...)
			}
			if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
				s.Assert().Equal(tc.expectedRecords, actualRecords, "resulting QuarantineRecords")
			}

			actualSends := bKeeper.SentCoins
			s.Assert().Equal(expectedSends, actualSends, "sends made")

			actualEvents := ctx.EventManager().Events()
			s.Assert().Equal(expectedEvents, actualEvents, "events emitted during accept")
		})
	}

	s.Run("send returns an error", func() {
		// Setup: There will be 4 records to send, the 3rd will return an error.
		// Check that:
		// 1. The error is returned by AcceptQuarantinedFunds
		// 2. The 1st and 2nd records are removed but the 3rd and 4th remain.
		// 3. SendCoins was called for the 1st and 2nd records.
		// 4. Events were emitted for the 1st and 2nd records.

		// Setup address stuff.
		addrBase := "sre"
		s.Require().False(seenAddrBases[addrBase], "an earlier test already used the address base %q", addrBase)
		seenAddrBases[addrBase] = true

		toAddr := MakeTestAddr(addrBase, 0)
		fromAddr1 := MakeTestAddr(addrBase, 1)
		fromAddr2 := MakeTestAddr(addrBase, 2)
		fromAddr3 := MakeTestAddr(addrBase, 3)
		fromAddr4 := MakeTestAddr(addrBase, 4)
		fromAddrs := []sdk.AccAddress{fromAddr1, fromAddr2, fromAddr3, fromAddr4}

		// Define the existing records and expected stuff.
		existingRecords := []*quarantine.QuarantineRecord{
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr1},
				Coins:                   cz("1addra"),
			},
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr2},
				Coins:                   cz("2addrb"),
			},
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr3},
				Coins:                   cz("3addrc"),
			},
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr4},
				Coins:                   cz("4addrd"),
			},
		}

		expectedErr := "this is a test error"

		expectedRecords := []*quarantine.QuarantineRecord{
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr3},
				Coins:                   cz("3addrc"),
			},
			{
				UnacceptedFromAddresses: []sdk.AccAddress{fromAddr4},
				Coins:                   cz("4addrd"),
			},
		}

		expectedSends := []*SentCoins{
			{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   toAddr,
				Amt:      cz("1addra"),
			},
			{
				FromAddr: s.keeper.GetFundsHolder(),
				ToAddr:   toAddr,
				Amt:      cz("2addrb"),
			},
		}

		expectedEvents := sdk.Events{
			makeEvent(s.T(), toAddr, cz("1addra")),
			makeEvent(s.T(), toAddr, cz("2addrb")),
		}

		// mock the bank keeper and set it to return an error on the 3rd send.
		bKeeper := NewMockBankKeeper() // bzzzzzzzzzz
		bKeeper.QueuedSendCoinsErrors = []error{
			nil,
			nil,
			fmt.Errorf(expectedErr),
		}
		qKeeper := s.keeper.WithBankKeeper(bKeeper)

		// Store the existing records.
		for i, record := range existingRecords {
			testFuncSet := func() {
				qKeeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
			}
			s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
		}

		// Do the thing.
		var actualErr error
		ctx := s.sdkCtx.WithEventManager(sdk.NewEventManager())
		testFuncAccept := func() {
			actualErr = qKeeper.AcceptQuarantinedFunds(ctx, toAddr, fromAddrs...)
		}
		s.Require().NotPanics(testFuncAccept, "AcceptQuarantinedFunds")

		// Check that: 1. The error is returned by AcceptQuarantinedFunds
		s.Assert().EqualError(actualErr, expectedErr, "AcceptQuarantinedFunds error")

		// Check that: 2. The 1st and 2nd records are removed but the 3rd and 4th remain.
		var actualRecords []*quarantine.QuarantineRecord
		testFuncGet := func() {
			actualRecords = qKeeper.GetQuarantineRecords(ctx, toAddr, fromAddrs...)
		}
		if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
			s.Assert().Equal(expectedRecords, actualRecords)
		}

		// Check that: 3. SendCoins was called for the 1st and 2nd records.
		actualSends := bKeeper.SentCoins
		s.Assert().Equal(expectedSends, actualSends, "sends made")

		// Check that: 4. Events were emitted for the 1st and 2nd records.
		actualEvents := ctx.EventManager().Events()
		s.Assert().Equal(expectedEvents, actualEvents, "events emitted")
	})
}

func (s *TestSuite) TestDeclineQuarantinedFunds() {
	// cz an even shorter way of creating coins since all creating should get the same *testing.T here.
	cz := func(coins string) sdk.Coins {
		return czt(s.T(), coins)
	}

	tests := []struct {
		name      string
		addrBase  string
		addrCount uint8
		fromAddrs []int
		existing  []*quarantine.QuarantineRecord
		expected  []*quarantine.QuarantineRecord
	}{
		{
			name:      "one from zero records",
			addrBase:  "ofzr",
			addrCount: 2,
			fromAddrs: []int{1},
			existing:  nil,
			expected:  nil,
		},
		{
			name:      "one from one record",
			addrBase:  "ofor",
			addrCount: 2,
			fromAddrs: []int{1},
			existing: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("13ofor"),
					Declined:                false,
				},
			},
			expected: []*quarantine.QuarantineRecord{
				{
					UnacceptedFromAddresses: []sdk.AccAddress{{1}},
					Coins:                   cz("13ofor"),
					Declined:                true,
				},
			},
		},
	}

	// Test cases:
	// one from one record previously accepted
	// one from two records
	// two froms zero records
	// two froms one record from first
	// two froms one record from second
	// two froms one record from both
	// two froms two records from first
	// two froms two records from second
	// two froms two records one from each
	// two froms two records one from one other from both
	// two froms five records (1st, 2nd, 1st & 2nd, 1st & other, 2nd & other)

	seenAddrBases := map[string]bool{}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Make sure the address base isn't used by an earlier test.
			s.Require().NotEqual(tc.addrBase, "", "no AddrBase defined")
			s.Require().False(seenAddrBases[tc.addrBase], "an earlier test already used the address base %q", tc.addrBase)
			seenAddrBases[tc.addrBase] = true
			s.Require().GreaterOrEqual(int(tc.addrCount), 1, "addrCount")

			// Set up all the address stuff.
			addrs := make([]sdk.AccAddress, tc.addrCount)
			for i := range addrs {
				addrs[i] = MakeTestAddr(tc.addrBase, uint8(i))
			}

			toAddr := addrs[0]
			fromAddrs := make([]sdk.AccAddress, len(tc.fromAddrs))
			for i, fi := range tc.fromAddrs {
				fromAddrs[i] = addrs[fi]
			}

			for _, record := range tc.existing {
				updateQR(addrs, record)
			}
			for _, record := range tc.expected {
				updateQR(addrs, record)
			}

			// Set the existing records.
			for i, record := range tc.existing {
				testFuncSet := func() {
					s.keeper.SetQuarantineRecord(s.sdkCtx, toAddr, record)
				}
				s.Require().NotPanics(testFuncSet, "SetQuarantineRecord[%d]", i)
			}

			// Do the thing.
			testFuncDecline := func() {
				s.keeper.DeclineQuarantinedFunds(s.sdkCtx, toAddr, fromAddrs...)
			}
			s.Require().NotPanics(testFuncDecline, "DeclineQuarantinedFunds")

			var actual []*quarantine.QuarantineRecord
			testFuncGet := func() {
				actual = s.keeper.GetQuarantineRecords(s.sdkCtx, toAddr, fromAddrs...)
			}
			if s.Assert().NotPanics(testFuncGet, "GetQuarantineRecords") {
				s.Assert().Equal(tc.expected, actual, "resulting quarantine records")
			}
		})
	}
}

// TODO[1046]: IterateQuarantineRecords
// TODO[1046]: GetAllQuarantinedFunds

// TODO[1046]: setQuarantineRecordSuffixIndex
// TODO[1046]: bzToQuarantineRecordSuffixIndex
// TODO[1046]: getQuarantineRecordSuffixIndex
// TODO[1046]: getQuarantineRecordSuffixes
// TODO[1046]: addQuarantineRecordSuffixIndexes
// TODO[1046]: deleteQuarantineRecordSuffixIndexes
