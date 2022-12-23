package keeper_test

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite

	app       *simapp.SimApp
	sdkCtx    sdk.Context
	stdlibCtx context.Context
	keeper    keeper.Keeper
	govKeeper *MockGovKeeper

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
	s.govKeeper = NewMockGovKeeper()
	s.keeper = s.app.SanctionKeeper.WithGovKeeper(s.govKeeper)

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

func (s *TestSuite) TestKeeper_GetAuthority() {
	s.Run("default", func() {
		expected := authtypes.NewModuleAddress(govtypes.ModuleName).String()
		actual := s.keeper.GetAuthority()
		s.Assert().Equal(expected, actual, "GetAuthority result")
	})

	tests := []string{"something", "something else"}
	for _, tc := range tests {
		s.Run(tc, func() {
			k := s.keeper.WithAuthority(tc)
			actual := k.GetAuthority()
			s.Assert().Equal(tc, actual, "GetAuthority result")
		})
	}
}

func (s *TestSuite) TestIsSanctionedAddr() {
	// Setup:
	// addr1 will be sanctioned.
	// addr2 will be sanctioned, but have a temp unsanction.
	// addr3 will have a temp sanction.
	// addr4 will have a temp sanction then temp unsanction.
	// addr5 will be sanctioned and have a temp unsanction then a temp sanction.
	// addrUnsanctionable will have a sanction in place, but be one of the unsanctionable addresses.
	var setupErr error
	addrUnsanctionable := sdk.AccAddress("unsanctionable_addr_")
	s.Require().NotPanics(func() {
		setupErr = s.keeper.SanctionAddresses(s.sdkCtx, s.addr1, s.addr2, s.addr5, addrUnsanctionable)
	}, "SanctionAddresses")
	s.Require().NoError(setupErr, "SanctionAddresses error")
	s.Require().NotPanics(func() {
		setupErr = s.keeper.AddTemporarySanction(s.sdkCtx, 1, s.addr3, s.addr4)
	}, "first AddTemporarySanction")
	s.Require().NoError(setupErr, "first AddTemporarySanction error")
	s.Require().NotPanics(func() {
		setupErr = s.keeper.AddTemporaryUnsanction(s.sdkCtx, 2, s.addr2, s.addr4, s.addr5)
	}, "AddTemporaryUnsanction")
	s.Require().NoError(setupErr, "AddTemporaryUnsanction error")
	s.Require().NotPanics(func() {
		setupErr = s.keeper.AddTemporarySanction(s.sdkCtx, 3, s.addr5)
	}, "second AddTemporarySanction")
	s.Require().NoError(setupErr, "second AddTemporarySanction error")

	k := s.keeper.WithUnsanctionableAddrs(map[string]bool{string(addrUnsanctionable): true})

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{
			name: "nil",
			addr: nil,
			exp:  false,
		},
		{
			name: "empty",
			addr: sdk.AccAddress{},
			exp:  false,
		},
		{
			name: "unknown address",
			addr: sdk.AccAddress("an__unknown__address"),
			exp:  false,
		},
		{
			name: "sanctioned addr",
			addr: s.addr1,
			exp:  true,
		},
		{
			name: "sanctioned with temp unsanction",
			addr: s.addr2,
			exp:  false,
		},
		{
			name: "temp sanction",
			addr: s.addr3,
			exp:  true,
		},
		{
			name: "temp sanction then temp unsanction",
			addr: s.addr4,
			exp:  false,
		},
		{
			name: "sanctioned with temp unsanction then temp sanction",
			addr: s.addr5,
			exp:  true,
		},
		{
			name: "first byte of sanctioned addr",
			addr: sdk.AccAddress{s.addr1[0]},
			exp:  false,
		},
		{
			name: "sanctioned addr plus 1 byte at end",
			addr: append(append([]byte{}, s.addr1...), 'f'),
			exp:  false,
		},
		{
			name: "sanctioned addr plus 1 byte at front",
			addr: append([]byte{'g'}, s.addr1...),
			exp:  false,
		},
		{
			name: "sanctioned addr minus last byte",
			addr: s.addr1[:len(s.addr1)-1],
			exp:  false,
		},
		{
			name: "sanctioned addr minus first byte",
			addr: s.addr1[1:],
			exp:  false,
		},
		{
			name: "sanctioned addr that is now unsanctionable",
			addr: addrUnsanctionable,
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var origAddr, origAtFullCap sdk.AccAddress
			if tc.addr != nil {
				origAddr = make(sdk.AccAddress, len(tc.addr), cap(tc.addr))
				copy(origAddr, tc.addr[:cap(tc.addr)])
				origAtFullCap = tc.addr[:cap(tc.addr)]
			}
			var actual bool
			testFunc := func() {
				actual = k.IsSanctionedAddr(s.sdkCtx, tc.addr)
			}
			s.Require().NotPanics(testFunc, "IsSanctionedAddr")
			s.Assert().Equal(tc.exp, actual, "IsSanctionedAddr result")
			s.Assert().Equal(origAddr, tc.addr, "provided addr before and after")
			s.Assert().Equal(cap(origAddr), cap(tc.addr), "provided addr capacity before and after")
			var addrAtFullCap sdk.AccAddress
			if tc.addr != nil {
				addrAtFullCap = tc.addr[:cap(tc.addr)]
			}
			s.Assert().Equal(origAtFullCap, addrAtFullCap, "provided addr at full capacity before and after")
		})
	}
}

// TODO[1046]: SanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error
// TODO[1046]: UnsanctionAddresses(ctx sdk.Context, addrs ...sdk.AccAddress) error
// TODO[1046]: AddTemporarySanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error
// TODO[1046]: AddTemporaryUnsanction(ctx sdk.Context, govPropID uint64, addrs ...sdk.AccAddress) error
// TODO[1046]: addTempEntries(ctx sdk.Context, value byte, govPropID uint64, addrs []sdk.AccAddress) error
// TODO[1046]: getLatestTempEntry(store sdk.KVStore, addr sdk.AccAddress) []byte
// TODO[1046]: DeleteGovPropTempEntries(ctx sdk.Context, govPropID uint64)
// TODO[1046]: DeleteAddrTempEntries(ctx sdk.Context, addrs ...sdk.AccAddress)
// TODO[1046]: IterateSanctionedAddresses(ctx sdk.Context, cb func(addr sdk.AccAddress) (stop bool))
// TODO[1046]: IterateTemporaryEntries(ctx sdk.Context, addr sdk.AccAddress, cb func(addr sdk.AccAddress, govPropID uint64, isSanction bool) (stop bool))
// TODO[1046]: IterateProposalIndexEntries(ctx sdk.Context, govPropID *uint64, cb func(govPropID uint64, addr sdk.AccAddress) (stop bool))

func (s *TestSuite) TestIsAddrThatCannotBeSanctioned() {
	k := s.keeper.WithUnsanctionableAddrs(map[string]bool{
		string(s.addr1): true,
		string(s.addr2): true,
		string(s.addr3): false, // I'm not sure how this would happen, but whatever.
	})

	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  bool
	}{
		{
			name: "unsanctionable addr 1",
			addr: s.addr1,
			exp:  true,
		},
		{
			name: "unsanctionable addr 2",
			addr: s.addr2,
			exp:  true,
		},
		{
			name: "sanctionable addr",
			addr: s.addr3,
			exp:  false,
		},
		{
			name: "nil",
			addr: nil,
			exp:  false,
		},
		{
			name: "empty",
			addr: nil,
			exp:  false,
		},
		{
			name: "random",
			addr: sdk.AccAddress("random"),
			exp:  false,
		},
		{
			name: "other addr",
			addr: s.addr5,
			exp:  false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual bool
			testFunc := func() {
				actual = k.IsAddrThatCannotBeSanctioned(tc.addr)
			}
			s.Require().NotPanics(testFunc, "IsAddrThatCannotBeSanctioned")
			s.Assert().Equal(tc.exp, actual, "IsAddrThatCannotBeSanctioned result")
		})
	}
}

func (s *TestSuite) TestGetSetParams() {
	// Change the defaults from their norm so we know they've got values we can check against.
	origSanct := sanction.DefaultImmediateSanctionMinDeposit
	origUnsanct := sanction.DefaultImmediateUnsanctionMinDeposit
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origSanct
		sanction.DefaultImmediateUnsanctionMinDeposit = origUnsanct
	}()
	sanction.DefaultImmediateSanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("sanct", 93))
	sanction.DefaultImmediateUnsanctionMinDeposit = sdk.NewCoins(sdk.NewInt64Coin("usanct", 49))

	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NotPanics(func() {
		s.keeper.DeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
	}, "DeleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
	s.Require().NotPanics(func() {
		s.keeper.DeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
	}, "DeleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)

	s.Run("get with no entries in store", func() {
		expected := sanction.DefaultParams()
		var actual *sanction.Params
		testGet := func() {
			actual = s.keeper.GetParams(s.sdkCtx)
		}
		s.Require().NotPanics(testGet, "GetParams")
		s.Assert().Equal(expected, actual, "GetParams result")
	})

	tests := []struct {
		name      string
		setInput  *sanction.Params
		getOutput *sanction.Params
	}{
		{
			name: "params with nils",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "empty coins",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.Coins{},
				ImmediateUnsanctionMinDeposit: sdk.Coins{},
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "only sanction",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sanct", 66)),
				ImmediateUnsanctionMinDeposit: nil,
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sanct", 66)),
				ImmediateUnsanctionMinDeposit: nil,
			},
		},
		{
			name: "only unsanction",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("unsuns", 5555)),
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   nil,
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("unsuns", 5555)),
			},
		},
		{
			name: "with both",
			setInput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sss", 123)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("uuu", 456)),
			},
			getOutput: &sanction.Params{
				ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin("sss", 123)),
				ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin("uuu", 456)),
			},
		},
		{
			name:      "nil",
			setInput:  nil,
			getOutput: sanction.DefaultParams(),
		},
	}

	paramsUpdatedEvent, eventErr := sdk.TypedEventToEvent(&sanction.EventParamsUpdated{})
	s.Require().NoError(eventErr, "sdk.TypedEventToEvent(&sanction.EventParamsUpdated{})")
	expectedEvents := sdk.Events{paramsUpdatedEvent}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.sdkCtx.WithEventManager(em)
			var err error
			testSet := func() {
				err = s.keeper.SetParams(ctx, tc.setInput)
			}
			s.Require().NotPanics(testSet, "SetParams")
			s.Require().NoError(err, "SetParams error")
			actualEvents := em.Events()
			s.Assert().Equal(expectedEvents, actualEvents, "events emitted during SetParams")
			var actual *sanction.Params
			testGet := func() {
				actual = s.keeper.GetParams(ctx)
			}
			s.Require().NotPanics(testGet, "GetParams")
			if !s.Assert().Equal(tc.getOutput, actual, "GetParams result") {
				if actual != nil {
					// it failed, but the coins aren't easy to read in that ouptput, so be helpful here.
					s.Assert().Equal(tc.getOutput.ImmediateSanctionMinDeposit.String(),
						actual.ImmediateSanctionMinDeposit.String(), "ImmediateSanctionMinDeposit")
					s.Assert().Equal(tc.getOutput.ImmediateUnsanctionMinDeposit.String(),
						actual.ImmediateUnsanctionMinDeposit.String(), "ImmediateUnsanctionMinDeposit")
				}
			}
		})
	}
}

func (s *TestSuite) TestIterateParams() {
	type kvPair struct {
		key   string
		value string
	}

	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	s.Require().NotPanics(func() {
		s.keeper.DeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
	}, "DeleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
	s.Require().NotPanics(func() {
		s.keeper.DeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
	}, "DeleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)

	s.Run("no entries", func() {
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return false
		}
		testFunc := func() {
			s.keeper.IterateParams(s.sdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Empty(actual, "params iterated")
	})

	// They should be iterated in alphabetical order by key, so they're ordered as such here.
	expected := []kvPair{
		{key: "param1", value: "value for param1"},
		{key: "param2", value: "param2 value"},
		{key: "param3", value: "the param3 value"},
		{key: "param4", value: "This is param4's value."},
		{key: "param5", value: "5valuecoin"},
	}
	// Write them in reverse order from expected.
	for i := len(expected) - 1; i >= 0; i-- {
		s.Require().NotPanics(func() {
			s.keeper.SetParam(store, expected[i].key, expected[i].value)
		}, "SetParam(%q, %q)", expected[i].key, expected[i].value)
	}

	s.Run("full iteration", func() {
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return false
		}
		testFunc := func() {
			s.keeper.IterateParams(s.sdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(expected, actual, "params iterated")
	})

	s.Run("stop after 3", func() {
		exp := []kvPair{expected[0], expected[1], expected[2]}
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return len(actual) >= len(exp)
		}
		testFunc := func() {
			s.keeper.IterateParams(s.sdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(exp, actual, "params iterated")
	})

	s.Run("stop after 1", func() {
		exp := []kvPair{expected[0]}
		var actual []kvPair
		cb := func(name, value string) bool {
			actual = append(actual, kvPair{key: name, value: value})
			return len(actual) >= len(exp)
		}
		testFunc := func() {
			s.keeper.IterateParams(s.sdkCtx, cb)
		}
		s.Require().NotPanics(testFunc, "IterateParams")
		s.Assert().Equal(exp, actual, "params iterated")
	})
}

func (s *TestSuite) TestGetImmediateSanctionMinDeposit() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// Set the defaults to different things to help sus out problems.
	origS := sanction.DefaultImmediateSanctionMinDeposit
	origU := sanction.DefaultImmediateUnsanctionMinDeposit
	sanction.DefaultImmediateSanctionMinDeposit = cz("3dflts")
	sanction.DefaultImmediateUnsanctionMinDeposit = cz("6dfltu")
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origS
		sanction.DefaultImmediateUnsanctionMinDeposit = origU
	}()

	// prep is something that should be done at the start of a test case.
	type prep struct {
		value  string
		set    bool
		delete bool
	}

	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	testFuncSetSanct := func() {
		s.keeper.SetParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit, "98unsanct")
	}
	s.Require().NotPanics(testFuncSetSanct, "SetParam(%q, %q)", keeper.ParamNameImmediateUnsanctionMinDeposit, "98unsanct")

	tests := []struct {
		name string
		prep []prep
		exp  sdk.Coins
	}{
		{
			name: "not in store",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
		{
			name: "empty string in store",
			prep: []prep{{value: "", set: true}},
			exp:  nil,
		},
		{
			name: "3sanct in store",
			prep: []prep{{value: "3sanct", set: true}},
			exp:  cz("3sanct"),
		},
		{
			name: "bad value in store",
			prep: []prep{{value: "how how", set: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
		{
			name: "not in store again",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateSanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, p := range tc.prep {
				if p.set {
					testFuncSet := func() {
						s.keeper.SetParam(store, keeper.ParamNameImmediateSanctionMinDeposit, p.value)
					}
					s.Require().NotPanics(testFuncSet, "SetParam(%q, %q)", keeper.ParamNameImmediateSanctionMinDeposit, p.value)
				}
				if p.delete {
					testFuncDelete := func() {
						s.keeper.DeleteParam(store, keeper.ParamNameImmediateSanctionMinDeposit)
					}
					s.Require().NotPanics(testFuncDelete, "DeleteParam(%q)", keeper.ParamNameImmediateSanctionMinDeposit)
				}
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.keeper.GetImmediateSanctionMinDeposit(s.sdkCtx)
			}
			s.Require().NotPanics(testFunc, "GetImmediateSanctionMinDeposit")
			s.Assert().Equal(tc.exp, actual, "GetImmediateSanctionMinDeposit result")
		})
	}
}

func (s *TestSuite) TestGetImmediateUnsanctionMinDeposit() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	// Set the defaults to different things to help sus out problems.
	origS := sanction.DefaultImmediateSanctionMinDeposit
	origU := sanction.DefaultImmediateUnsanctionMinDeposit
	sanction.DefaultImmediateSanctionMinDeposit = cz("2dflts")
	sanction.DefaultImmediateUnsanctionMinDeposit = cz("5dfltu")
	defer func() {
		sanction.DefaultImmediateSanctionMinDeposit = origS
		sanction.DefaultImmediateUnsanctionMinDeposit = origU
	}()

	// prep is something that should be done at the start of a test case.
	type prep struct {
		value  string
		set    bool
		delete bool
	}

	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	testFuncSetSanct := func() {
		s.keeper.SetParam(store, keeper.ParamNameImmediateSanctionMinDeposit, "99sanct")
	}
	s.Require().NotPanics(testFuncSetSanct, "SetParam(%q, %q)", keeper.ParamNameImmediateSanctionMinDeposit, "99sanct")

	tests := []struct {
		name string
		prep []prep
		exp  sdk.Coins
	}{
		{
			name: "not in store",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
		{
			name: "empty string in store",
			prep: []prep{{value: "", set: true}},
			exp:  nil,
		},
		{
			name: "3unsanct in store",
			prep: []prep{{value: "3unsanct", set: true}},
			exp:  cz("3unsanct"),
		},
		{
			name: "bad value in store",
			prep: []prep{{value: "what what", set: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
		{
			name: "not in store again",
			prep: []prep{{delete: true}},
			exp:  sanction.DefaultImmediateUnsanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, p := range tc.prep {
				if p.set {
					testFuncSet := func() {
						s.keeper.SetParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit, p.value)
					}
					s.Require().NotPanics(testFuncSet, "SetParam(%q, %q)", keeper.ParamNameImmediateUnsanctionMinDeposit, p.value)
				}
				if p.delete {
					testFuncDelete := func() {
						s.keeper.DeleteParam(store, keeper.ParamNameImmediateUnsanctionMinDeposit)
					}
					s.Require().NotPanics(testFuncDelete, "DeleteParam(%q)", keeper.ParamNameImmediateUnsanctionMinDeposit)
				}
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.keeper.GetImmediateUnsanctionMinDeposit(s.sdkCtx)
			}
			s.Require().NotPanics(testFunc, "GetImmediateUnsanctionMinDeposit")
			s.Assert().Equal(tc.exp, actual, "GetImmediateUnsanctionMinDeposit result")
		})
	}
}

func (s *TestSuite) TestGetSetDeleteParam() {
	store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
	var toDelete []string

	newParamName := "new param"
	s.Run("get param that does not exist", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "GetParam(%q)", newParamName)
		s.Assert().Equal("", actual, "GetParam(%q) result string", newParamName)
		s.Assert().False(ok, "GetParam(%q) result bool", newParamName)
	})

	newParamValue := "new param value"
	s.Run("set param new param", func() {
		var alreadyExists bool
		testFuncGet := func() {
			_, alreadyExists = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "GetParam(%q) on param that should not exist yet", newParamName)
		s.Require().False(alreadyExists, "GetParam(%q) result bool on param that should not exist yet", newParamName)
		testFuncSet := func() {
			s.keeper.SetParam(store, newParamName, newParamValue)
		}
		s.Require().NotPanics(testFuncSet, "SetParam(%q, %q)", newParamName, newParamValue)
		toDelete = append(toDelete, newParamName)
	})

	s.Run("get param new param", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "GetParam(%q)", newParamName)
		s.Require().True(ok, "GetParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "GetParam(%q) result string", newParamName)
	})

	s.Run("set and get fruits", func() {
		name := "fruits"
		value := "bananas, apples, pears, papaya, pineapple, pomegranate"
		testFuncSet := func() {
			s.keeper.SetParam(store, name, value)
		}
		s.Require().NotPanics(testFuncSet, "SetParam(%q, %q)", name, value)
		toDelete = append(toDelete, name)
		var actual string
		var ok bool
		testfuncGet := func() {
			actual, ok = s.keeper.GetParam(store, name)
		}
		s.Require().NotPanics(testfuncGet, "GetParam(%q)", name)
		s.Assert().True(ok, "GetParam(%q) result bool", name)
		s.Assert().Equal(value, actual, "GetParam(%q) result string", name)
	})

	s.Run("get new param again", func() {
		var actual string
		var ok bool
		testFuncGet := func() {
			actual, ok = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet, "GetParam(%q)", newParamName)
		s.Require().True(ok, "GetParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "GetParam(%q) result string", newParamName)
	})

	s.Run("update and get first param", func() {
		var alreadyExists bool
		testFuncGet1 := func() {
			_, alreadyExists = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet1, "GetParam(%q) on param that should not exist yet", newParamName)
		s.Require().True(alreadyExists, "GetParam(%q) result bool on param that should not exist yet", newParamName)
		newParamValue = "this is an updated new param value"
		testFuncSet := func() {
			s.keeper.SetParam(store, newParamName, newParamValue)
		}
		s.Require().NotPanics(testFuncSet, "SetParam(%q, %q)", newParamName, newParamValue)

		var actual string
		var ok bool
		testFuncGet2 := func() {
			actual, ok = s.keeper.GetParam(store, newParamName)
		}
		s.Require().NotPanics(testFuncGet2, "GetParam(%q)", newParamName)
		s.Require().True(ok, "GetParam(%q) result bool", newParamName)
		s.Require().Equal(newParamValue, actual, "GetParam(%q) result string", newParamName)
	})

	for _, name := range toDelete {
		s.Run("delete "+name, func() {
			testDeleteFunc := func() {
				s.keeper.DeleteParam(store, name)
			}
			s.Require().NotPanics(testDeleteFunc, "DeleteParam(%q)", name)
			var actual string
			var ok bool
			testGetFunc := func() {
				actual, ok = s.keeper.GetParam(store, name)
			}
			s.Require().NotPanics(testGetFunc, "GetParam(%q)", name)
			s.Assert().False(ok, "GetParam(%q) result bool", name)
			s.Assert().Equal("", actual, "GetParam(%q) result string", name)
		})
	}

	s.Run("delete new param again", func() {
		testDeleteFunc := func() {
			s.keeper.DeleteParam(store, newParamName)
		}
		s.Require().NotPanics(testDeleteFunc, "DeleteParam(%q)", newParamName)
	})
}

func (s *TestSuite) TestGetParamAsCoinsOrDefault() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}

	tests := []struct {
		name     string
		setFirst bool
		setTo    string
		param    string
		dflt     sdk.Coins
		exp      sdk.Coins
	}{
		{
			name:     "unknown name",
			setFirst: false,
			param:    "unknown",
			dflt:     cz("1default"),
			exp:      cz("1default"),
		},
		{
			name:     "param not a coin",
			setFirst: true,
			setTo:    "not a coin",
			param:    "not-a-coin",
			dflt:     cz("1default"),
			exp:      cz("1default"),
		},
		{
			name:     "empty string",
			setFirst: true,
			setTo:    "",
			param:    "empty-string",
			dflt:     cz("1default"),
			exp:      nil,
		},
		{
			name:     "coin string one denom",
			setFirst: true,
			setTo:    "5acoin",
			param:    "one-denom",
			dflt:     cz("1default"),
			exp:      cz("5acoin"),
		},
		{
			name:     "coin string two denoms",
			setFirst: true,
			setTo:    "4acoin,10walnut",
			param:    "two-denom",
			dflt:     cz("1default"),
			exp:      cz("4acoin,10walnut"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.setFirst {
				store := s.sdkCtx.KVStore(s.keeper.GetStoreKey())
				s.keeper.SetParam(store, tc.param, tc.setTo)
				defer func() {
					s.keeper.DeleteParam(store, tc.param)
				}()
			}
			var actual sdk.Coins
			testFunc := func() {
				actual = s.keeper.GetParamAsCoinsOrDefault(s.sdkCtx, tc.param, tc.dflt)
			}
			s.Require().NotPanics(testFunc, "GetParamAsCoinsOrDefault")
			s.Assert().Equal(tc.exp, actual, "GetParamAsCoinsOrDefault result")
		})
	}
}

func (s *TestSuite) TestToCoinsOrDefault() {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		s.Require().NoError(err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	tests := []struct {
		name  string
		coins string
		dflt  sdk.Coins
		exp   sdk.Coins
	}{
		{
			name:  "empty string",
			coins: "",
			dflt:  cz("1defaultcoin,2banana"),
			exp:   nil,
		},
		{
			name:  "bad string",
			coins: "bad",
			dflt:  cz("1goodcoin,8defaults"),
			exp:   cz("1goodcoin,8defaults"),
		},
		{
			name:  "one denom",
			coins: "1particle",
			dflt:  cz("8quark"),
			exp:   cz("1particle"),
		},
		{
			name:  "two denoms",
			coins: "50handcoin,99gloves",
			dflt:  cz("42towels"),
			exp:   cz("50handcoin,99gloves"),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual sdk.Coins
			testFunc := func() {
				actual = keeper.ToCoinsOrDefault(tc.coins, tc.dflt)
			}
			s.Require().NotPanics(testFunc, "ToCoinsOrDefault")
			s.Assert().Equal(tc.exp, actual, "ToCoinsOrDefault result")
		})
	}
}

func (s *TestSuite) TestToAccAddrs() {
	tests := []struct {
		name   string
		addrs  []string
		exp    []sdk.AccAddress
		expErr []string
	}{
		{
			name:  "nil list",
			addrs: nil,
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "empty list",
			addrs: []string{},
			exp:   []sdk.AccAddress{},
		},
		{
			name:  "one good address",
			addrs: []string{sdk.AccAddress("one good address").String()},
			exp:   []sdk.AccAddress{sdk.AccAddress("one good address")},
		},
		{
			name:   "one bad address",
			addrs:  []string{"one bad address"},
			expErr: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "five addresses all good",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			exp: []sdk.AccAddress{
				sdk.AccAddress("good address 0"),
				sdk.AccAddress("good address 1"),
				sdk.AccAddress("good address 2"),
				sdk.AccAddress("good address 3"),
				sdk.AccAddress("good address 4"),
			},
		},
		{
			name: "five addresses first bad",
			addrs: []string{
				"bad address 0",
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			expErr: []string{"invalid address[0]", "decoding bech32 failed"},
		},
		{
			name: "five addresses third bad",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				"bad address 2",
				sdk.AccAddress("good address 3").String(),
				sdk.AccAddress("good address 4").String(),
			},
			expErr: []string{"invalid address[2]", "decoding bech32 failed"},
		},
		{
			name: "five addresses fifth bad",
			addrs: []string{
				sdk.AccAddress("good address 0").String(),
				sdk.AccAddress("good address 1").String(),
				sdk.AccAddress("good address 2").String(),
				sdk.AccAddress("good address 3").String(),
				"bad address 4",
			},
			expErr: []string{"invalid address[4]", "decoding bech32 failed"},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			var actual []sdk.AccAddress
			var err error
			testFunc := func() {
				actual, err = keeper.ToAccAddrs(tc.addrs)
			}
			s.Require().NotPanics(testFunc, "toAccAddrs")
			testutil.AssertErrorContents(s.T(), err, tc.expErr, "toAccAddrs error")
			s.Assert().Equal(tc.exp, actual, "toAccAddrs result")
		})
	}
}
