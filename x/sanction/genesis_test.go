package sanction_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

func TestNewGenesisState(t *testing.T) {
	cz := func(coins string) sdk.Coins {
		rv, err := sdk.ParseCoinsNormalized(coins)
		require.NoError(t, err, "ParseCoinsNormalized(%q)", coins)
		return rv
	}
	tests := []struct {
		name   string
		params *sanction.Params
		addrs  []string
		exp    *sanction.GenesisState
	}{
		{
			name:   "nil nil",
			params: nil,
			addrs:  nil,
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: nil,
			},
		},
		{
			name:   "nil empty",
			params: nil,
			addrs:  []string{},
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: []string{},
			},
		},
		{
			name:   "empty nil",
			params: &sanction.Params{},
			addrs:  nil,
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{},
				SanctionedAddresses: nil,
			},
		},
		{
			name:   "empty empty",
			params: &sanction.Params{},
			addrs:  []string{},
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{},
				SanctionedAddresses: []string{},
			},
		},
		{
			name:   "only-sanct-dep nil",
			params: &sanction.Params{ImmediateSanctionMinDeposit: cz("5sanct")},
			addrs:  nil,
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{ImmediateSanctionMinDeposit: cz("5sanct")},
				SanctionedAddresses: nil,
			},
		},
		{
			name:   "only-unsanct-dep nil",
			params: &sanction.Params{ImmediateUnsanctionMinDeposit: cz("8usanct")},
			addrs:  nil,
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{ImmediateUnsanctionMinDeposit: cz("8usanct")},
				SanctionedAddresses: nil,
			},
		},
		{
			name: "both params nil addrs",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   cz("11sanct"),
				ImmediateUnsanctionMinDeposit: cz("13usanct"),
			},
			addrs: nil,
			exp: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("11sanct"),
					ImmediateUnsanctionMinDeposit: cz("13usanct"),
				},
				SanctionedAddresses: nil,
			},
		},
		{
			name:   "nil params 3 addrs",
			params: nil,
			addrs:  []string{"addr1", "addr2", "addr3"},
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: []string{"addr1", "addr2", "addr3"},
			},
		},
		{
			name: "a little of both",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   cz("11sanct"),
				ImmediateUnsanctionMinDeposit: cz("13usanct"),
			},
			addrs: []string{"addr-one", "addr-two", "addr-three", "addr-fourteen"}, // Bono, why?
			exp: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("11sanct"),
					ImmediateUnsanctionMinDeposit: cz("13usanct"),
				},
				SanctionedAddresses: []string{"addr-one", "addr-two", "addr-three", "addr-fourteen"},
			},
		},
		{
			name:   "default",
			params: sanction.DefaultParams(),
			addrs:  nil,
			exp:    sanction.DefaultGenesisState(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual *sanction.GenesisState
			testFunc := func() {
				actual = sanction.NewGenesisState(tc.params, tc.addrs)
			}
			require.NotPanics(t, testFunc, "NewGenesisState")
			if assert.NotNil(t, actual, "NewGenesisState result") {
				if !assert.Equal(t, tc.exp, actual, "NewGenesisState result") {
					// If we get here, at least one of these should fail and hopefully help point to the thing that's different.
					assert.Equal(t, tc.exp.Params, actual.Params, "NewGenesisState Params")
					assert.Equal(t, tc.exp.SanctionedAddresses, actual.SanctionedAddresses, "NewGenesisState SanctionedAddresses")
				}
			}
		})
	}
}

func TestGenesisState_Validate(t *testing.T) {

	tests := []struct {
		name string
		gs   *sanction.GenesisState
		exp  []string
	}{
		{
			name: "empty genesis state",
			gs:   &sanction.GenesisState{},
			exp:  nil,
		},
		{
			name: "invalid params",
			gs: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   sdk.Coins{sdk.NewInt64Coin("dcoin", 1), sdk.NewInt64Coin("dcoin", 2)},
					ImmediateUnsanctionMinDeposit: nil,
				},
				SanctionedAddresses: nil,
			},
			exp: []string{"invalid params", "invalid immediate sanction min deposit", "duplicate denomination dcoin"},
		},
		{
			name: "invalid first address",
			gs: &sanction.GenesisState{
				Params: nil,
				SanctionedAddresses: []string{
					"not1avalidaddr0",
					sdk.AccAddress("testaddr1___________").String(),
					sdk.AccAddress("testaddr2___________").String(),
					sdk.AccAddress("testaddr3___________").String(),
					sdk.AccAddress("testaddr4___________").String(),
				},
			},
			exp: []string{"invalid address", "sanctioned addresses[0]", "decoding bech32 failed", `"not1avalidaddr0"`},
		},
		{
			name: "invalid third address",
			gs: &sanction.GenesisState{
				Params: nil,
				SanctionedAddresses: []string{
					sdk.AccAddress("testaddr0___________").String(),
					sdk.AccAddress("testaddr1___________").String(),
					"not1avalidaddr2",
					sdk.AccAddress("testaddr3___________").String(),
					sdk.AccAddress("testaddr4___________").String(),
				},
			},
			exp: []string{"invalid address", "sanctioned addresses[2]", "decoding bech32 failed", `"not1avalidaddr2"`},
		},
		{
			name: "invalid last address",
			gs: &sanction.GenesisState{
				Params: nil,
				SanctionedAddresses: []string{
					sdk.AccAddress("testaddr0___________").String(),
					sdk.AccAddress("testaddr1___________").String(),
					sdk.AccAddress("testaddr2___________").String(),
					sdk.AccAddress("testaddr3___________").String(),
					"not1avalidaddr4",
				},
			},
			exp: []string{"invalid address", "sanctioned addresses[4]", "decoding bech32 failed", `"not1avalidaddr4"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			testFunc := func() {
				err = tc.gs.Validate()
			}
			require.NotPanics(t, testFunc, "GenesisState.Validate()")
			testutil.AssertErrorContents(t, err, tc.exp, ".Validate result")
		})
	}
}
