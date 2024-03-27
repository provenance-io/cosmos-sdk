package keeper_test

import (
	"context"
	"strings"
	"time"

	"github.com/golang/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// reqFundAcc calls FundAccount and requires that it doesn't return an error.
func (suite *KeeperTestSuite) reqFundAcc(addr sdk.AccAddress, amt sdk.Coins) {
	suite.mockFundAccount(addr)
	suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, addr, amt), "FundAccount(%q, %q)", string(addr), amt)
}

// resetLockedCoinsFnDeferrable creates a deferable function that resets the
// bank keeper's locked coins getter to what it was originally.
// Usage: defer resetLockedCoinsFnDeferrable()()
// The second () causes this function to get called at the defer line so that
// when it runs at the end, it's running the result of that call.
func (suite *KeeperTestSuite) resetLockedCoinsFnDeferrable() func() {
	origLockedCoinsGetter := suite.bankKeeper.GetLockedCoinsGetter()
	return func() {
		suite.bankKeeper.SetLockedCoinsGetter(origLockedCoinsGetter)
	}
}

func (suite *KeeperTestSuite) TestUnvestedCoins() {
	defer suite.resetLockedCoinsFnDeferrable()()

	addrNoVest := sdk.AccAddress("addrNoVest__________")
	addrContVest := sdk.AccAddress("addrConVest_________")
	addrPerVest := sdk.AccAddress("addrPerVest_________")
	addrDelVest := sdk.AccAddress("addrDelVest_________")
	addrPermLock := sdk.AccAddress("addrPermLock________")

	coins := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("fish", amt))
	}

	// Set up one normal account and one of each vesting account.
	// Each applicable account will have 1,000,000 fish vesting ending in 1,000,000 seconds.
	// Each will also have an additional 1,000 fish just to make sure it doesn't matter.
	vestBal := coins(1_000_000)
	addlBal := coins(1_000)
	totalBal := vestBal.Add(addlBal...)
	startTime := time.Unix(1700000000, 0) // chosen for the even number. It's 2023-11-14 22:13:20 (a Tuesday).
	endTime := startTime.Add(1_000_000 * time.Second)

	var accs []sdk.AccountI
	accNoVest := authtypes.NewBaseAccountWithAddress(addrNoVest)
	suite.reqFundAcc(addrNoVest, totalBal)
	accs = append(accs, accNoVest)

	// Continuous vesting account for 1,000,000 fish over 1,000,000 seconds (11.6ish days).
	accContVest, err := vesting.NewContinuousVestingAccount(
		authtypes.NewBaseAccountWithAddress(addrContVest), vestBal, startTime.Unix(), endTime.Unix())
	suite.Require().NoError(err, "NewContinuousVestingAccount(%q, %q, %s, %s)",
		string(addrContVest), vestBal, startTime.String(), endTime.String())
	suite.reqFundAcc(addrContVest, totalBal)
	accs = append(accs, accContVest)

	// A periodic vesting account that vests 250,000 fish every 250,000 seconds four times.
	accPerVest, err := vesting.NewPeriodicVestingAccount(
		authtypes.NewBaseAccountWithAddress(addrPerVest), vestBal, startTime.Unix(), vesting.Periods{
			{Length: 250_000, Amount: coins(250_000)},
			{Length: 250_000, Amount: coins(250_000)},
			{Length: 250_000, Amount: coins(250_000)},
			{Length: 250_000, Amount: coins(250_000)},
		},
	)
	suite.Require().NoError(err, "NewPeriodicVestingAccount(%q, %q, %s, ...)",
		string(addrPerVest), vestBal, startTime.String())
	suite.reqFundAcc(addrPerVest, totalBal)
	accs = append(accs, accPerVest)

	// A delayed vesting account that vests 1,000,000 fish after 1,000,000 seconds.
	accDelVest, err := vesting.NewDelayedVestingAccount(
		authtypes.NewBaseAccountWithAddress(addrDelVest), vestBal, endTime.Unix())
	suite.Require().NoError(err, "NewDelayedVestingAccount(%q, %q, %s)",
		string(addrDelVest), vestBal, endTime.String())
	suite.reqFundAcc(addrDelVest, totalBal)
	accs = append(accs, accDelVest)

	// A permanent locked account with 1,000,000 fish locked.
	accPermLock, err := vesting.NewPermanentLockedAccount(authtypes.NewBaseAccountWithAddress(addrPermLock), vestBal)
	suite.Require().NoError(err, "NewPermanentLockedAccount(%q, %q)", string(addrPermLock), vestBal)
	suite.reqFundAcc(addrPermLock, totalBal)
	accs = append(accs, accPermLock)

	for _, acc := range accs {
		suite.authKeeper.EXPECT().GetAccount(gomock.Any(), acc.GetAddress()).Return(acc).AnyTimes()
	}

	// This should be completely ignored in the UnvestedCoins function.
	alsoLocked := coins(50)
	suite.bankKeeper.AppendLockedCoinsGetter(func(_ context.Context, _ sdk.AccAddress) sdk.Coins {
		return alsoLocked
	})

	tests := []struct {
		name   string
		bypass bool
		time   time.Time
		addr   sdk.AccAddress
		expAmt sdk.Coins
	}{
		{
			name:   "normal account",
			time:   startTime,
			addr:   addrNoVest,
			expAmt: nil,
		},

		{
			name:   "continuous vesting at start",
			time:   startTime,
			addr:   addrContVest,
			expAmt: vestBal,
		},
		{
			name:   "continuous vesting at start with bypass",
			bypass: true,
			time:   startTime,
			addr:   addrContVest,
			expAmt: nil,
		},
		{
			name:   "continuous vesting after 250,000 seconds",
			time:   startTime.Add(250_000 * time.Second),
			addr:   addrContVest,
			expAmt: vestBal.Sub(coins(250_000)...),
		},
		{
			name:   "continuous vesting after 250,000 seconds with bypass",
			bypass: true,
			time:   startTime.Add(250_000 * time.Second),
			addr:   addrContVest,
			expAmt: nil,
		},
		{
			name:   "continuous vesting after 500,005 seconds",
			time:   startTime.Add(500_005 * time.Second),
			addr:   addrContVest,
			expAmt: vestBal.Sub(coins(500_005)...),
		},
		{
			name:   "continuous vesting after 500,005 seconds with bypass",
			bypass: true,
			time:   startTime.Add(500_005 * time.Second),
			addr:   addrContVest,
			expAmt: nil,
		},
		{
			name:   "continuous vesting after full term",
			time:   endTime,
			addr:   addrContVest,
			expAmt: nil,
		},

		{
			name:   "periodic vesting just before first period",
			time:   startTime.Add(249_999 * time.Second),
			addr:   addrPerVest,
			expAmt: vestBal,
		},
		{
			name:   "periodic vesting just before first period with bypass",
			bypass: true,
			time:   startTime.Add(249_999 * time.Second),
			addr:   addrPerVest,
			expAmt: nil,
		},
		{
			name:   "periodic vesting at first period",
			time:   startTime.Add(250_000 * time.Second),
			addr:   addrPerVest,
			expAmt: vestBal.Sub(coins(250_000)...),
		},
		{
			name:   "periodic vesting at second period",
			time:   startTime.Add(500_000 * time.Second),
			addr:   addrPerVest,
			expAmt: vestBal.Sub(coins(500_000)...),
		},
		{
			name:   "periodic vesting at third period",
			time:   startTime.Add(750_000 * time.Second),
			addr:   addrPerVest,
			expAmt: vestBal.Sub(coins(750_000)...),
		},
		{
			name:   "periodic vesting at end",
			time:   endTime,
			addr:   addrPerVest,
			expAmt: nil,
		},

		{
			name:   "delayed vesting at start",
			time:   startTime,
			addr:   addrDelVest,
			expAmt: vestBal,
		},
		{
			name:   "delayed vesting at start with bypass",
			bypass: true,
			time:   startTime,
			addr:   addrDelVest,
			expAmt: nil,
		},
		{
			name:   "delayed vesting just before end",
			time:   endTime.Add(-1 * time.Second),
			addr:   addrDelVest,
			expAmt: vestBal,
		},
		{
			name:   "delayed vesting at end",
			time:   endTime,
			addr:   addrDelVest,
			expAmt: nil,
		},

		{
			name:   "permanent locked at start",
			time:   startTime,
			addr:   addrPermLock,
			expAmt: vestBal,
		},
		{
			name:   "permanent locked at start with bypass",
			bypass: true,
			time:   startTime,
			addr:   addrPermLock,
			expAmt: nil,
		},
		{
			name:   "permanent locked much later",
			time:   endTime.Add(1_000_000_000 * time.Second),
			addr:   addrPermLock,
			expAmt: vestBal,
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx := sdk.UnwrapSDKContext(suite.ctx).WithBlockTime(tc.time)
			if tc.bypass {
				ctx = types.WithVestingLockedBypass(ctx)
			}

			var amt sdk.Coins
			testFunc := func() {
				amt = suite.bankKeeper.UnvestedCoins(ctx, tc.addr)
			}
			suite.Require().NotPanics(testFunc, "UnvestedCoins")
			suite.Assert().Equal(tc.expAmt.String(), amt.String(), "UnvestedCoins result")
		})
	}
}

func (suite *KeeperTestSuite) TestLockedCoins() {
	defer suite.resetLockedCoinsFnDeferrable()()

	// Set up two different locked coins getters and record their args.
	var locked1, locked2 sdk.Coins
	var ctx1, ctx2 context.Context
	var addr1, addr2 sdk.AccAddress
	suite.bankKeeper.AppendLockedCoinsGetter(func(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
		ctx1 = ctx
		addr1 = addr
		return locked1
	})
	suite.bankKeeper.AppendLockedCoinsGetter(func(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
		ctx2 = ctx
		addr2 = addr
		return locked2
	})

	customKey := "custom-ctx-test-key"
	withCustomValue := func(ctx context.Context, value string) context.Context {
		return sdk.UnwrapSDKContext(ctx).WithValue(customKey, value)
	}
	getCustomValue := func(ctx context.Context) string {
		valI := ctx.Value(customKey)
		if valI != nil {
			val, ok := valI.(string)
			if ok {
				return val
			}
		}
		return ""
	}

	coins := func(amt string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(amt)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", amt)
		return rv
	}

	// Create a permanent locked vesting account
	addrPermLock := sdk.AccAddress("addrPermLock________")
	accPermLockBase := authtypes.NewBaseAccountWithAddress(addrPermLock)
	permLockBal := coins("100000stones")
	accPermLock, err := vesting.NewPermanentLockedAccount(accPermLockBase, permLockBal)
	suite.Require().NoError(err, "NewPermanentLockedAccount(%q, %q)", string(addrPermLock), permLockBal)
	addlBal := coins("500stones")
	suite.reqFundAcc(addrPermLock, permLockBal.Add(addlBal...))
	suite.authKeeper.EXPECT().GetAccount(gomock.Any(), addrPermLock).Return(accPermLock).AnyTimes()

	tests := []struct {
		name    string
		vBypass bool
		locked1 sdk.Coins
		locked2 sdk.Coins
		addr    sdk.AccAddress
		expAmt  sdk.Coins
	}{
		{
			name:    "nothing locked",
			locked1: nil,
			locked2: nil,
			expAmt:  nil,
		},
		{
			name:    "only an amount from first",
			locked1: coins("3rocks"),
			locked2: nil,
			expAmt:  coins("3rocks"),
		},
		{
			name:    "only an amount from second",
			locked1: nil,
			locked2: coins("12rocks"),
			expAmt:  coins("12rocks"),
		},
		{
			name:    "same denom from both",
			locked1: coins("50rocks"),
			locked2: coins("300rocks"),
			expAmt:  coins("350rocks"),
		},
		{
			name:    "different denoms from each",
			locked1: coins("50rocks"),
			locked2: coins("7000pebbles"),
			expAmt:  coins("7000pebbles,50rocks"),
		},
		{
			name:    "perm locked plus some",
			locked1: coins("83rocks"),
			locked2: coins("12stones"),
			addr:    addrPermLock,
			expAmt:  coins("83rocks,12stones").Add(permLockBal...),
		},
		{
			name:    "perm locked plus some with vesting bypass",
			vBypass: true,
			locked1: coins("83rocks"),
			locked2: coins("12stones"),
			addr:    addrPermLock,
			expAmt:  coins("83rocks,12stones"),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			ctx1, ctx2, addr1, addr2 = nil, nil, nil, nil
			locked1 = tc.locked1
			locked2 = tc.locked2

			if len(tc.addr) == 0 {
				tc.addr = sdk.AccAddress(tc.name)
				suite.authKeeper.EXPECT().GetAccount(gomock.Any(), tc.addr).Return(nil).AnyTimes()
			}
			ctx := withCustomValue(suite.ctx, tc.name)
			if tc.vBypass {
				ctx = types.WithVestingLockedBypass(ctx)
			}
			var amt sdk.Coins
			testFunc := func() {
				amt = suite.bankKeeper.LockedCoins(ctx, tc.addr)
			}
			suite.Require().NotPanics(testFunc, "LockedCoins(sdk.AccAddress(%q))", string(tc.addr))

			suite.Assert().Equal(tc.expAmt.String(), amt.String(), "LockedCoins result")
			suite.Assert().Equal(tc.addr, addr1, "addr given to first getter")
			suite.Assert().Equal(tc.addr, addr2, "addr given to second getter")
			ctx1Val := getCustomValue(ctx1)
			suite.Assert().Equal(tc.name, ctx1Val, "custom context value in first getter")
			ctx2Val := getCustomValue(ctx2)
			suite.Assert().Equal(tc.name, ctx2Val, "custom context value in second getter")
		})
	}
}

func (suite *KeeperTestSuite) TestLockedCoins_SpendableCoins() {
	defer suite.resetLockedCoinsFnDeferrable()()

	suite.bankKeeper.ClearLockedCoinsGetter()
	var locked sdk.Coins
	suite.bankKeeper.AppendLockedCoinsGetter(func(_ context.Context, _ sdk.AccAddress) sdk.Coins {
		return locked
	})

	coins := func(amt string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(amt)
		suite.Require().NoError(err, "ParseCoinsNormalized(%q)", amt)
		return rv
	}

	// Mock the mint account for when we fund the accounts. We can't just use mockFundAccount (or reqFundAcc)
	// because we've taken the UnvestedCoins out of the picture, so there's no longer that call to GetAccount.
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, mintAcc.Name).Return(mintAcc).AnyTimes()
	suite.authKeeper.EXPECT().GetModuleAddress(mintAcc.Name).Return(mintAcc.GetAddress()).AnyTimes()

	tests := []struct {
		name    string
		balance sdk.Coins
		locked  sdk.Coins
		expAmt  sdk.Coins
	}{
		{
			name:    "no balance no locked",
			balance: nil,
			locked:  nil,
			expAmt:  nil,
		},
		{
			name:    "one denom none locked",
			balance: coins("88banana"),
			locked:  nil,
			expAmt:  coins("88banana"),
		},
		{
			name:    "one denom all locked",
			balance: coins("88banana"),
			locked:  coins("88banana"),
			expAmt:  nil,
		},
		{
			name:    "one denom some locked",
			balance: coins("88banana"),
			locked:  coins("3banana"),
			expAmt:  coins("85banana"),
		},
		{
			name:    "locked denom not in account",
			balance: coins("88banana"),
			locked:  coins("12acorn"),
			expAmt:  coins("88banana"),
		},
		{
			name:    "one denom more locked than available",
			balance: coins("88banana"),
			locked:  coins("99banana"),
			expAmt:  nil,
		},
		{
			name:    "two denoms none locked",
			balance: coins("12acorn,88banana"),
			locked:  nil,
			expAmt:  coins("12acorn,88banana"),
		},
		{
			name:    "two denoms all locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("12acorn,88banana"),
			expAmt:  nil,
		},
		{
			name:    "two denoms all of one locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("12acorn"),
			expAmt:  coins("88banana"),
		},
		{
			name:    "two denoms too much of first locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("14acorn,15banana"),
			expAmt:  coins("73banana"),
		},
		{
			name:    "two denoms too much of second locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("4acorn,99banana"),
			expAmt:  coins("8acorn"),
		},
		{
			name:    "two denoms a little of each locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("1acorn,2banana"),
			expAmt:  coins("11acorn,86banana"),
		},
		{
			name:    "two balance denoms three locked",
			balance: coins("12acorn,88banana"),
			locked:  coins("1acorn,2banana,3corn"),
			expAmt:  coins("11acorn,86banana"),
		},
	}

	usedAddrs := make(map[string]bool)

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			addr := sdk.AccAddress(strings.ReplaceAll(tc.name, " ", "_") + strings.Repeat("_", 20))[:32]
			// Make sure we haven't used this address in a previous test (which would throw things off).
			suite.Require().False(usedAddrs[string(addr)], "has addr %s been used", string(addr))
			usedAddrs[string(addr)] = true

			locked = nil // Otherwise FundAccount thinks there's locked things for it's send.
			suite.authKeeper.EXPECT().HasAccount(suite.ctx, addr).Return(true)
			suite.Require().NoError(banktestutil.FundAccount(suite.ctx, suite.bankKeeper, addr, tc.balance),
				"FundAccount(%q, %q)", string(addr), tc.balance)

			// Fund's over. Lock it down.
			locked = tc.locked
			var amt sdk.Coins
			testFunc := func() {
				amt = suite.bankKeeper.SpendableCoins(suite.ctx, addr)
			}
			suite.Require().NotPanics(testFunc, "SpendableCoins(%q)", string(addr))
			suite.Assert().Equal(tc.expAmt.String(), amt.String(), "SpendableCoins result")
		})
	}
}

func (suite *KeeperTestSuite) TestLockedCoins_SendCoins() {
	defer suite.resetLockedCoinsFnDeferrable()()
	// This is just a simple test to make sure the SendCoins function pays attention to locked coins.
	fromAddr := sdk.AccAddress("just_a_from_address_")
	toAddr := sdk.AccAddress("just_a_to_address___")
	balance := sdk.NewCoins(sdk.NewInt64Coin("acorn", 12), sdk.NewInt64Coin("banana", 99))
	locked := sdk.NewCoins(sdk.NewInt64Coin("acorn", 5), sdk.NewInt64Coin("banana", 27))
	toSend := sdk.NewCoins(sdk.NewInt64Coin("acorn", 8))
	suite.reqFundAcc(fromAddr, balance)

	suite.bankKeeper.ClearLockedCoinsGetter()
	suite.bankKeeper.AppendLockedCoinsGetter(func(_ context.Context, addr sdk.AccAddress) sdk.Coins {
		if addr.Equals(fromAddr) {
			return locked
		}
		return nil
	})

	expErr := "spendable balance 7acorn is smaller than 8acorn: insufficient funds"
	var err error
	testFunc := func() {
		err = suite.bankKeeper.SendCoins(suite.ctx, fromAddr, toAddr, toSend)
	}
	suite.Require().NotPanics(testFunc, "SendCoins")
	suite.Assert().EqualError(err, expErr, "SendCoins error")
}

func (suite *KeeperTestSuite) TestLockedCoins_InputOutputCoins() {
	defer suite.resetLockedCoinsFnDeferrable()()
	// This is just a simple test to make sure the InputOutputCoins function pays attention to locked coins.
	fromAddr := sdk.AccAddress("just_a_from_address_")
	toAddr := sdk.AccAddress("just_a_to_address___")
	balance := sdk.NewCoins(sdk.NewInt64Coin("acorn", 12), sdk.NewInt64Coin("banana", 99))
	locked := sdk.NewCoins(sdk.NewInt64Coin("acorn", 5), sdk.NewInt64Coin("banana", 27))
	toSend := sdk.NewCoins(sdk.NewInt64Coin("acorn", 8), sdk.NewInt64Coin("banana", 32))
	suite.reqFundAcc(fromAddr, balance)

	suite.bankKeeper.ClearLockedCoinsGetter()
	suite.bankKeeper.AppendLockedCoinsGetter(func(_ context.Context, addr sdk.AccAddress) sdk.Coins {
		if addr.Equals(fromAddr) {
			return locked
		}
		return nil
	})

	expErr := "spendable balance 7acorn is smaller than 8acorn: insufficient funds"
	inputs := []types.Input{{Address: fromAddr.String(), Coins: toSend}}
	outputs := []types.Output{{Address: toAddr.String(), Coins: toSend}}
	var err error
	testFunc := func() {
		err = suite.bankKeeper.InputOutputCoins(suite.ctx, inputs[0], outputs)
	}
	suite.Require().NotPanics(testFunc, "InputOutputCoins")
	suite.Assert().ErrorContains(err, expErr, "InputOutputCoins error")
}

func (suite *KeeperTestSuite) TestLockedCoins_DelegateCoins() {
	suite.resetLockedCoinsFnDeferrable()()
	// This makes sure that the DelegateCoins ignores coins locked due
	// to vesting, but not other locked coins.

	coins := func(amt int64) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin("stones", amt))
	}

	// The module address doesn't actually have to be a module for this.
	addrModule := sdk.AccAddress("addrModule__________")
	accModule := authtypes.NewBaseAccountWithAddress(addrModule)
	suite.authKeeper.EXPECT().GetAccount(gomock.Any(), addrModule).Return(accModule).AnyTimes()

	addrPermLock := sdk.AccAddress("addrPermLock________")
	accPermLockBase := authtypes.NewBaseAccountWithAddress(addrPermLock)
	permLockFunds := coins(100_000)
	accPermLock, err := vesting.NewPermanentLockedAccount(accPermLockBase, permLockFunds)
	suite.Require().NoError(err, "NewPermanentLockedAccount(%q, %q)", string(addrPermLock), permLockFunds)
	addlFunds := coins(500)
	suite.reqFundAcc(addrPermLock, permLockFunds.Add(addlFunds...))
	suite.authKeeper.EXPECT().GetAccount(gomock.Any(), addrPermLock).Return(accPermLock).AnyTimes()

	otherLocked := coins(10)
	var hasBypass bool
	suite.bankKeeper.AppendLockedCoinsGetter(func(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
		hasBypass = types.HasVestingLockedBypass(ctx)
		if addr.Equals(addrPermLock) {
			return otherLocked
		}
		return nil
	})

	// All vesting can be delegated as well as the additional funds.
	// But the other locked funds can't be delegated.
	// So we try to delegate one more than that so that it fails, but would pass
	// if the other locked coins are ignored.
	toDelegate := coins(100_491)

	expErr := "failed to delegate; amount available 100490stones is less than 100491stones: insufficient funds"
	testFunc := func() {
		err = suite.bankKeeper.DelegateCoins(suite.ctx, addrPermLock, addrModule, toDelegate)
	}
	suite.Require().NotPanics(testFunc, "DelegateCoins")
	suite.Assert().EqualError(err, expErr, "DelegateCoins error")
	suite.Assert().True(hasBypass, "making sure the context provided to the locked coins function has the vesting bypass")
}
