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
