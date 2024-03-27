package types

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestVestingLockedContextFuncsSdk(t *testing.T) {
	sdkCtxMaker := func() sdk.Context {
		return sdk.NewContext(nil, cmtproto.Header{}, false, nil)
	}
	runVestingLockedContextFuncsTests(t, sdkCtxMaker)
}

func TestVestingLockedContextFuncsStdlib(t *testing.T) {
	stdlibCtxMaker := func() context.Context {
		return context.Context(sdk.NewContext(nil, cmtproto.Header{}, false, nil))
	}
	runVestingLockedContextFuncsTests(t, stdlibCtxMaker)
}

func runVestingLockedContextFuncsTests[C context.Context](t *testing.T, ctxMaker func() C) {
	tests := []struct {
		name       string
		ctxWrapper func(ctx C) C
		expHas     bool
	}{
		{
			name: "fresh context",
			ctxWrapper: func(ctx C) C {
				return ctx
			},
			expHas: false,
		},
		{
			name:       "context with bypass",
			ctxWrapper: WithVestingLockedBypass[C],
			expHas:     true,
		},
		{
			name: "context with bypass on one that originally was without it",
			ctxWrapper: func(ctx C) C {
				return WithVestingLockedBypass(WithoutVestingLockedBypass(ctx))
			},
			expHas: true,
		},
		{
			name: "context with bypass twice",
			ctxWrapper: func(ctx C) C {
				return WithVestingLockedBypass(WithVestingLockedBypass(ctx))
			},
			expHas: true,
		},
		{
			name:       "context without bypass",
			ctxWrapper: WithoutVestingLockedBypass[C],
			expHas:     false,
		},
		{
			name: "context without bypass on one that originally had it",
			ctxWrapper: func(ctx C) C {
				return WithoutVestingLockedBypass(WithVestingLockedBypass(ctx))
			},
			expHas: false,
		},
		{
			name: "context without bypass twice",
			ctxWrapper: func(ctx C) C {
				return WithoutVestingLockedBypass(WithoutVestingLockedBypass(ctx))
			},
			expHas: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := ctxMaker()
			wrapFunc := func() {
				ctx = tc.ctxWrapper(ctx)
			}
			require.NotPanics(t, wrapFunc, "ctxWrapper")
			var actHas bool
			testFunc := func() {
				actHas = HasVestingLockedBypass(ctx)
			}
			require.NotPanics(t, testFunc, "HasVestingLockedBypass")
			assert.Equal(t, tc.expHas, actHas, "HasVestingLockedBypass")
		})
	}

	t.Run("does not modify provided", func(t *testing.T) {
		origCtx := ctxMaker()
		assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx)")
		afterWith := WithVestingLockedBypass(origCtx)
		assert.True(t, HasVestingLockedBypass(afterWith), "HasVestingLockedBypass(afterWith)")
		assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx) after giving it to WithVestingLockedBypass")
		afterWithout := WithoutVestingLockedBypass(afterWith)
		assert.False(t, HasVestingLockedBypass(afterWithout), "HasVestingLockedBypass(afterWithout)")
		assert.True(t, HasVestingLockedBypass(afterWith), "HasVestingLockedBypass(afterWith) after giving it to WithoutVestingLockedBypass")
		assert.False(t, HasVestingLockedBypass(origCtx), "HasVestingLockedBypass(origCtx) after giving afterWith to WithoutVestingLockedBypass")
	})
}

func TestKeyContainsSpecificName(t *testing.T) {
	assert.Contains(t, bypassKey, "vesting", "bypassKey")
	assert.Contains(t, bypassKey, "bypass", "bypassKey")
}

// GetLockedCoinsArgs are the args provided to a GetLockedCoinsFn function.
type GetLockedCoinsArgs struct {
	Name string
	Addr sdk.AccAddress
}

// GetLockedCoinsTestHelper is a struct with stuff helpful for testing the GetLockedCoinsFn stuff.
type GetLockedCoinsTestHelper struct {
	Calls []*GetLockedCoinsArgs
}

func NewGetLockedCoinsTestHelper() *GetLockedCoinsTestHelper {
	return &GetLockedCoinsTestHelper{Calls: make([]*GetLockedCoinsArgs, 0, 2)}
}

// RecordCall makes note that the provided args were used as a GetLockedCoinsFn call.
func (s *GetLockedCoinsTestHelper) RecordCall(name string, addr sdk.AccAddress) {
	s.Calls = append(s.Calls, s.NewArgs(name, addr))
}

// NewCalls is just a shorter way to create a []*GetLockedCoinsArgs.
func (s *GetLockedCoinsTestHelper) NewCalls(args ...*GetLockedCoinsArgs) []*GetLockedCoinsArgs {
	return args
}

// NewArgs creates a new GetLockedCoinsArgs.
func (s *GetLockedCoinsTestHelper) NewArgs(name string, addr sdk.AccAddress) *GetLockedCoinsArgs {
	return &GetLockedCoinsArgs{
		Name: name,
		Addr: addr,
	}
}

// NamedGetter creates a new GetLockedCoinsFn function that records the arguments it's called with and returns an amount.
func (s *GetLockedCoinsTestHelper) NamedGetter(name string, amt sdk.Coins) GetLockedCoinsFn {
	return func(_ context.Context, addr sdk.AccAddress) sdk.Coins {
		s.RecordCall(name, addr)
		return amt
	}
}

// GetLockedCoinsTestParams are parameters to test regarding calling a GetLockedCoinsFn.
type GetLockedCoinsTestParams struct {
	// ExpNil is whether to expect the provided GetLockedCoinsFn to be nil.
	// If it is true, the rest of these test params are ignored.
	ExpNil bool
	// Addr is the address to use as input.
	Addr sdk.AccAddress
	// ExpCoins is the expected output coins.
	ExpCoins sdk.Coins
	// ExpCalls is the args of all the GetLockedCoinsFn calls that end up being made.
	ExpCalls []*GetLockedCoinsArgs
}

// TestActual tests the provided GetLockedCoinsFn using the provided test parameters.
func (s *GetLockedCoinsTestHelper) TestActual(t *testing.T, tp *GetLockedCoinsTestParams, actual GetLockedCoinsFn) {
	t.Helper()
	if tp.ExpNil {
		require.Nil(t, actual, "resulting GetLockedCoinsFn")
	} else {
		require.NotNil(t, actual, "resulting GetLockedCoinsFn")
		s.Calls = s.Calls[:0]
		lockedCoins := actual(sdk.Context{}, tp.Addr)
		assert.Equal(t, tp.ExpCoins.String(), lockedCoins.String(), "composit GetLockedCoinsFn output coins")
		assert.Equal(t, tp.ExpCalls, s.Calls, "args given to funcs in composit GetLockedCoinsFn")
	}
}

func TestGetLockedCoins_Then(t *testing.T) {
	addr := sdk.AccAddress("addr________________")
	cz := func(amount int64, denom string) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amount))
	}
	cz2 := func(amount1 int64, denom1 string, amount2 int64, denom2 string) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom1, amount1), sdk.NewInt64Coin(denom2, amount2))
	}

	h := NewGetLockedCoinsTestHelper()

	tests := []struct {
		name   string
		base   GetLockedCoinsFn
		second GetLockedCoinsFn
		exp    *GetLockedCoinsTestParams
	}{

		{
			name:   "nil nil",
			base:   nil,
			second: nil,
			exp: &GetLockedCoinsTestParams{
				ExpNil: true,
			},
		},
		{
			name:   "nil noop empty",
			base:   nil,
			second: h.NamedGetter("noop", sdk.Coins{}),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:   "nil noop nil",
			base:   nil,
			second: h.NamedGetter("noop", nil),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: nil,
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:   "noop nil",
			base:   h.NamedGetter("noop", sdk.Coins{}),
			second: nil,
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:   "noop noop",
			base:   h.NamedGetter("noop1", sdk.Coins{}),
			second: h.NamedGetter("noop2", sdk.Coins{}),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop1", addr), h.NewArgs("noop2", addr)),
			},
		},
		{
			name:   "two with same denoms",
			base:   h.NamedGetter("1acoin", cz(1, "acoin")),
			second: h.NamedGetter("2acoin", cz(2, "acoin")),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz(3, "acoin"),
				ExpCalls: h.NewCalls(h.NewArgs("1acoin", addr), h.NewArgs("2acoin", addr)),
			},
		},
		{
			name:   "two with different denoms",
			base:   h.NamedGetter("acoin", cz(1, "acoin")),
			second: h.NamedGetter("bcoin", cz(2, "bcoin")),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz2(1, "acoin", 2, "bcoin"),
				ExpCalls: h.NewCalls(h.NewArgs("acoin", addr), h.NewArgs("bcoin", addr)),
			},
		},
		{
			name:   "double chain",
			base:   ComposeGetLockedCoins(h.NamedGetter("r1", cz(1, "foo")), h.NamedGetter("r2", cz(2, "bar"))),
			second: ComposeGetLockedCoins(h.NamedGetter("r3", cz(4, "foo")), h.NamedGetter("r4", cz(8, "bar"))),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz2(5, "foo", 10, "bar"),
				ExpCalls: h.NewCalls(
					h.NewArgs("r1", addr),
					h.NewArgs("r2", addr),
					h.NewArgs("r3", addr),
					h.NewArgs("r4", addr),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual GetLockedCoinsFn
			testFunc := func() {
				actual = tc.base.Then(tc.second)
			}
			require.NotPanics(t, testFunc, "GetLockedCoinsFn.Then")
			h.TestActual(t, tc.exp, actual)
		})
	}
}

func TestComposeGetLockedCoins(t *testing.T) {
	addr := sdk.AccAddress("________addr________")
	fnz := func(rs ...GetLockedCoinsFn) []GetLockedCoinsFn {
		return rs
	}
	cz := func(amount int64, denom string) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom, amount))
	}
	cz2 := func(amount1 int64, denom1 string, amount2 int64, denom2 string) sdk.Coins {
		return sdk.NewCoins(sdk.NewInt64Coin(denom1, amount1), sdk.NewInt64Coin(denom2, amount2))
	}

	h := NewGetLockedCoinsTestHelper()

	tests := []struct {
		name  string
		input []GetLockedCoinsFn
		exp   *GetLockedCoinsTestParams
	}{
		{
			name:  "nil list",
			input: nil,
			exp: &GetLockedCoinsTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "empty list",
			input: fnz(),
			exp: &GetLockedCoinsTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "only nil entry",
			input: fnz(nil),
			exp: &GetLockedCoinsTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "five nil entries",
			input: fnz(nil, nil, nil, nil, nil),
			exp: &GetLockedCoinsTestParams{
				ExpNil: true,
			},
		},
		{
			name:  "only noop entry",
			input: fnz(h.NamedGetter("noop", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:  "only one entry",
			input: fnz(h.NamedGetter("acorns", cz(99, "acoin"))),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz(99, "acoin"),
				ExpCalls: h.NewCalls(h.NewArgs("acorns", addr)),
			},
		},
		{
			name:  "noop nil nil",
			input: fnz(h.NamedGetter("noop", sdk.Coins{}), nil, nil),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:  "nil noop nil",
			input: fnz(nil, h.NamedGetter("noop", sdk.Coins{}), nil),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:  "nil nil noop",
			input: fnz(nil, nil, h.NamedGetter("noop", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("noop", addr)),
			},
		},
		{
			name:  "noop noop nil",
			input: fnz(h.NamedGetter("r1", sdk.Coins{}), h.NamedGetter("r2", sdk.Coins{}), nil),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("r2", addr)),
			},
		},
		{
			name:  "noop nil noop",
			input: fnz(h.NamedGetter("r1", sdk.Coins{}), nil, h.NamedGetter("r2", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("r2", addr)),
			},
		},
		{
			name:  "nil noop noop",
			input: fnz(nil, h.NamedGetter("r1", sdk.Coins{}), h.NamedGetter("r2", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("r2", addr)),
			},
		},
		{
			name:  "noop noop noop",
			input: fnz(h.NamedGetter("r1", sdk.Coins{}), h.NamedGetter("r2", sdk.Coins{}), h.NamedGetter("r3", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: sdk.Coins{},
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("r2", addr), h.NewArgs("r3", addr)),
			},
		},
		{
			name:  "1acoin noop noop",
			input: fnz(h.NamedGetter("acorns", cz(1, "acoin")), h.NamedGetter("r2", sdk.Coins{}), h.NamedGetter("r3", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz(1, "acoin"),
				ExpCalls: h.NewCalls(h.NewArgs("acorns", addr), h.NewArgs("r2", addr), h.NewArgs("r3", addr)),
			},
		},
		{
			name:  "noop 2acoin noop",
			input: fnz(h.NamedGetter("r1", sdk.Coins{}), h.NamedGetter("acorns", cz(2, "acoin")), h.NamedGetter("r3", sdk.Coins{})),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz(2, "acoin"),
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("acorns", addr), h.NewArgs("r3", addr)),
			},
		},
		{
			name:  "noop noop 3acoin",
			input: fnz(h.NamedGetter("r1", sdk.Coins{}), h.NamedGetter("r2", sdk.Coins{}), h.NamedGetter("acorns", cz(3, "acoin"))),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz(3, "acoin"),
				ExpCalls: h.NewCalls(h.NewArgs("r1", addr), h.NewArgs("r2", addr), h.NewArgs("acorns", addr)),
			},
		},
		{
			name:  "1acoin 2bcoin 4acoin",
			input: fnz(h.NamedGetter("acorns", cz(1, "acoin")), h.NamedGetter("bcorns", cz(2, "bcoin")), h.NamedGetter("not sea corns", cz(4, "acoin"))),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz2(5, "acoin", 2, "bcoin"),
				ExpCalls: h.NewCalls(h.NewArgs("acorns", addr), h.NewArgs("bcorns", addr), h.NewArgs("not sea corns", addr)),
			},
		},
		{
			name: "big bang",
			input: fnz(nil,
				h.NamedGetter("noop0", nil), nil,
				h.NamedGetter("g1", cz(1, "bananas")), nil, nil,
				h.NamedGetter("noop1", sdk.Coins{}),
				h.NamedGetter("noop2", sdk.Coins{}),
				h.NamedGetter("g2", cz2(98, "bananas", 5, "apples")), nil,
				h.NamedGetter("g3", cz(8, "apples")), nil,
				h.NamedGetter("noop3", sdk.Coins{}),
			),
			exp: &GetLockedCoinsTestParams{
				Addr:     addr,
				ExpCoins: cz2(99, "bananas", 13, "apples"),
				ExpCalls: h.NewCalls(
					h.NewArgs("noop0", addr),
					h.NewArgs("g1", addr),
					h.NewArgs("noop1", addr),
					h.NewArgs("noop2", addr),
					h.NewArgs("g2", addr),
					h.NewArgs("g3", addr),
					h.NewArgs("noop3", addr),
				),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual GetLockedCoinsFn
			testFunc := func() {
				actual = ComposeGetLockedCoins(tc.input...)
			}
			require.NotPanics(t, testFunc, "ComposeGetLockedCoins")
			h.TestActual(t, tc.exp, actual)
		})
	}
}

func TestNoOpGetLockedCoinsFn(t *testing.T) {
	var lockedCoins sdk.Coins
	testFunc := func() {
		lockedCoins = NoOpGetLockedCoinsFn(sdk.Context{}, sdk.AccAddress{})
	}
	require.NotPanics(t, testFunc, "NoOpGetLockedCoinsFn")
	if assert.NotNil(t, lockedCoins, "NoOpGetLockedCoinsFn coins") {
		assert.Equal(t, "", lockedCoins.String(), "NoOpGetLockedCoinsFn coins")
	}
}
