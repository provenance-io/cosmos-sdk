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
		temps  []*sanction.TemporaryEntry
		exp    *sanction.GenesisState
	}{
		{
			name:   "nil nil",
			params: nil,
			addrs:  nil,
			temps:  nil,
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name:   "nil empty empty",
			params: nil,
			addrs:  []string{},
			temps:  []*sanction.TemporaryEntry{},
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: []string{},
				TemporaryEntries:    []*sanction.TemporaryEntry{},
			},
		},
		{
			name:   "empty nil empty",
			params: &sanction.Params{},
			addrs:  nil,
			temps:  []*sanction.TemporaryEntry{},
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{},
				SanctionedAddresses: nil,
				TemporaryEntries:    []*sanction.TemporaryEntry{},
			},
		},
		{
			name:   "empty empty empty",
			params: &sanction.Params{},
			addrs:  []string{},
			temps:  []*sanction.TemporaryEntry{},
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{},
				SanctionedAddresses: []string{},
				TemporaryEntries:    []*sanction.TemporaryEntry{},
			},
		},
		{
			name:   "only-sanct-dep nil nil",
			params: &sanction.Params{ImmediateSanctionMinDeposit: cz("5sanct")},
			addrs:  nil,
			temps:  nil,
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{ImmediateSanctionMinDeposit: cz("5sanct")},
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name:   "only-unsanct-dep nil nil",
			params: &sanction.Params{ImmediateUnsanctionMinDeposit: cz("8usanct")},
			addrs:  nil,
			temps:  nil,
			exp: &sanction.GenesisState{
				Params:              &sanction.Params{ImmediateUnsanctionMinDeposit: cz("8usanct")},
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name: "both params nil addrs and temps",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   cz("11sanct"),
				ImmediateUnsanctionMinDeposit: cz("13usanct"),
			},
			addrs: nil,
			temps: nil,
			exp: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("11sanct"),
					ImmediateUnsanctionMinDeposit: cz("13usanct"),
				},
				SanctionedAddresses: nil,
				TemporaryEntries:    nil,
			},
		},
		{
			name:   "nil params 3 addrs nil temps",
			params: nil,
			addrs:  []string{"addr1", "addr2", "addr3"},
			temps:  nil,
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: []string{"addr1", "addr2", "addr3"},
				TemporaryEntries:    nil,
			},
		},
		{
			name:   "nil params nil addrs 3 temps",
			params: nil,
			addrs:  nil,
			temps: []*sanction.TemporaryEntry{
				{
					Address:    "addr4",
					ProposalId: 4,
					Status:     sanction.TEMP_STATUS_SANCTIONED,
				},
				{
					Address:    "addr5",
					ProposalId: 5,
					Status:     sanction.TEMP_STATUS_UNSANCTIONED,
				},
				{
					Address:    "addr6",
					ProposalId: 6,
					Status:     8,
				},
			},
			exp: &sanction.GenesisState{
				Params:              nil,
				SanctionedAddresses: nil,
				TemporaryEntries: []*sanction.TemporaryEntry{
					{
						Address:    "addr4",
						ProposalId: 4,
						Status:     sanction.TEMP_STATUS_SANCTIONED,
					},
					{
						Address:    "addr5",
						ProposalId: 5,
						Status:     sanction.TEMP_STATUS_UNSANCTIONED,
					},
					{
						Address:    "addr6",
						ProposalId: 6,
						Status:     8,
					},
				},
			},
		},
		{
			name: "a little of all",
			params: &sanction.Params{
				ImmediateSanctionMinDeposit:   cz("11sanct"),
				ImmediateUnsanctionMinDeposit: cz("13usanct"),
			},
			addrs: []string{"addr-one", "addr-two", "addr-three", "addr-fourteen"}, // Bono, why?
			temps: []*sanction.TemporaryEntry{
				{
					Address:    "addr-twenty",
					ProposalId: 8,
					Status:     sanction.TEMP_STATUS_SANCTIONED,
				},
				{
					Address:    "addr-twenty-one",
					ProposalId: 9,
					Status:     sanction.TEMP_STATUS_UNSANCTIONED,
				},
			},
			exp: &sanction.GenesisState{
				Params: &sanction.Params{
					ImmediateSanctionMinDeposit:   cz("11sanct"),
					ImmediateUnsanctionMinDeposit: cz("13usanct"),
				},
				SanctionedAddresses: []string{"addr-one", "addr-two", "addr-three", "addr-fourteen"},
				TemporaryEntries: []*sanction.TemporaryEntry{
					{
						Address:    "addr-twenty",
						ProposalId: 8,
						Status:     sanction.TEMP_STATUS_SANCTIONED,
					},
					{
						Address:    "addr-twenty-one",
						ProposalId: 9,
						Status:     sanction.TEMP_STATUS_UNSANCTIONED,
					},
				},
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
				actual = sanction.NewGenesisState(tc.params, tc.addrs, tc.temps)
			}
			require.NotPanics(t, testFunc, "NewGenesisState")
			if assert.NotNil(t, actual, "NewGenesisState result") {
				if !assert.Equal(t, tc.exp, actual, "NewGenesisState result") {
					// If we get here, at least one of these should fail and hopefully help point to the thing that's different.
					assert.Equal(t, tc.exp.Params, actual.Params, "NewGenesisState Params")
					assert.Equal(t, tc.exp.SanctionedAddresses, actual.SanctionedAddresses, "NewGenesisState SanctionedAddresses")
					assert.Equal(t, tc.exp.TemporaryEntries, actual.TemporaryEntries, "NewGenesisState TemporaryEntries")
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
