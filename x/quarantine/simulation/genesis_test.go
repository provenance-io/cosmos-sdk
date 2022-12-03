package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/cosmos/cosmos-sdk/x/quarantine/simulation"
)

func TestRandomizedGenState(t *testing.T) {
	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          simapp.MakeTestEncodingConfig().Codec,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 10),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	var err error
	bankGenBefore := banktypes.GenesisState{}
	simState.GenState[banktypes.ModuleName], err = simState.Cdc.MarshalJSON(&bankGenBefore)
	require.NoError(t, err, "MarshalJSON empty bank genesis state")

	fundsHolder := authtypes.NewModuleAddress(quarantine.ModuleName)

	simulation.RandomizedGenState(&simState, fundsHolder)
	var gen quarantine.GenesisState
	err = simState.Cdc.UnmarshalJSON(simState.GenState[quarantine.ModuleName], &gen)
	require.NoError(t, err, "UnmarshalJSON on quarantine genesis state")

	totalQuarantined := sdk.Coins{}
	for _, qf := range gen.QuarantinedFunds {
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	if !totalQuarantined.IsZero() {
		var bankGen banktypes.GenesisState
		err = simState.Cdc.UnmarshalJSON(simState.GenState[banktypes.ModuleName], &bankGen)
		require.NoError(t, err, "UnmarshalJSON on quarantine bank state")
		holder := fundsHolder.String()
		holderFound := false
		for _, bal := range bankGen.Balances {
			if holder == bal.Address {
				holderFound = true
				assert.Equal(t, totalQuarantined.String(), bal.Coins.String())
			}
		}
		assert.True(t, holderFound, "no balance entry found for the funds holder")
		_, hasNeg := bankGen.Supply.SafeSub(totalQuarantined...)
		assert.False(t, hasNeg, "not enough supply %s to cover the total quarantined %s", bankGen.Supply.String(), totalQuarantined.String())
	}
}

func TestRandomQuarantinedAddresses(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// So all we can do here is check the length of the returned list using seed values found through trial and error.

	type testCase struct {
		name   string
		seed   int64
		expLen int
	}

	tests := []*testCase{
		{
			name:   "zero",
			seed:   103,
			expLen: 0,
		},
		{
			name:   "one",
			seed:   3,
			expLen: 1,
		},
		{
			name:   "two",
			seed:   17,
			expLen: 2,
		},
		{
			name:   "three",
			seed:   2,
			expLen: 3,
		},
		{
			name:   "four",
			seed:   4,
			expLen: 4,
		},
		{
			name:   "five",
			seed:   0,
			expLen: 5,
		},
		{
			name:   "six",
			seed:   15,
			expLen: 6,
		},
		{
			name:   "seven",
			seed:   31,
			expLen: 7,
		},
		{
			name:   "eight",
			seed:   45,
			expLen: 8,
		},
		{
			name:   "nine",
			seed:   238,
			expLen: 9,
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		// Using a separate rand to generate accounts to make it easier to predict the func being tested.
		accounts := simtypes.RandomAccounts(rand.New(rand.NewSource(1)), tc.expLen*4)
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantinedAddresses(r, accounts)
		if assert.Len(t, actual, tc.expLen, "QuarantinedAddresses") {
			if tc.expLen == 0 {
				rv = assert.Nil(t, actual, "QuarantinedAddresses") && rv
			}
		} else {
			rv = false
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for !runTest(t, tc) {
					tc.seed += 1
				}
			}
		})
		t.Run("good seeds", func(t *testing.T) {
			for _, tc := range tests {
				t.Logf("%d => %q", tc.seed, tc.name)
			}
			t.Fail()
		})
	}
}

func TestRandomQuarantineAutoResponses(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// In here, using seeds found through trial and error, we can check that some
	// addrs are kept, others ignored, and some new ones added.

	type testCase struct {
		name     string
		seed     int64
		qAddrs   []string
		expAddrs []string
		newAddrs int
	}

	tests := []*testCase{
		{
			name:     "no addrs in no new addrs",
			seed:     3,
			qAddrs:   nil,
			expAddrs: nil,
			newAddrs: 0,
		},
		{
			name:     "no addrs in one new addr",
			seed:     1,
			qAddrs:   nil,
			expAddrs: nil,
			newAddrs: 1,
		},
		{
			name:     "one addr in is kept",
			seed:     5,
			qAddrs:   []string{"addr1"},
			expAddrs: []string{"addr1"},
			newAddrs: 0,
		},
		{
			name:     "one addr in is not kept",
			seed:     4,
			qAddrs:   []string{"addr1"},
			expAddrs: nil,
			newAddrs: 0,
		},
		{
			name:     "two addrs in first kept new added",
			seed:     2,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr1"},
			newAddrs: 1,
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantineAutoResponses(r, tc.qAddrs)
		addrMap := make(map[string]bool)
		for _, entry := range actual {
			addrMap[entry.ToAddress] = true
		}
		addrs := make([]string, 0, len(addrMap))
		for addr := range addrMap {
			addrs = append(addrs, addr)
		}
		rv = assert.Len(t, addrs, len(tc.expAddrs)+tc.newAddrs, "to addresses") && rv
		for _, addr := range tc.expAddrs {
			rv = assert.Contains(t, addrs, addr, "to addresses") && rv
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for !runTest(t, tc) {
					tc.seed += 1
				}
			}
		})
		t.Run("good seeds", func(t *testing.T) {
			for _, tc := range tests {
				t.Logf("%d => %q", tc.seed, tc.name)
			}
			t.Fail()
		})
	}
}

func TestRandomQuarantinedFunds(t *testing.T) {
	// Once RandomAccounts is called, we can't trust the values returned from r.
	// In here, using seeds found through trial and error, we can check that some
	// addrs are kept, others ignored, and some new ones added.

	type testCase struct {
		name     string
		seed     int64
		qAddrs   []string
		expAddrs []string
	}

	tests := []*testCase{
		{
			name:     "no addrs in",
			seed:     0,
			qAddrs:   nil,
			expAddrs: nil,
		},
		{
			name:     "one addr in is kept",
			seed:     0,
			qAddrs:   []string{"addr1"},
			expAddrs: []string{"addr1"},
		},
		{
			name:     "one addr in is not kept",
			seed:     3,
			qAddrs:   []string{"addr1"},
			expAddrs: nil,
		},
		{
			name:     "two addrs in first kept",
			seed:     4,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr1"},
		},
		{
			name:     "two addrs in second kept",
			seed:     3,
			qAddrs:   []string{"addr1", "addr2"},
			expAddrs: []string{"addr2"},
		},
	}

	runTest := func(t *testing.T, tc *testCase) bool {
		t.Helper()
		rv := true
		r := rand.New(rand.NewSource(tc.seed))
		actual := simulation.RandomQuarantinedFunds(r, tc.qAddrs)
		addrMap := make(map[string]bool)
		for _, entry := range actual {
			addrMap[entry.ToAddress] = true
		}
		addrs := make([]string, 0, len(addrMap))
		for addr := range addrMap {
			addrs = append(addrs, addr)
		}
		rv = assert.Len(t, addrs, len(tc.expAddrs), "to addresses") && rv
		for _, addr := range tc.expAddrs {
			rv = assert.Contains(t, addrs, addr, "to addresses") && rv
		}
		return rv
	}

	allPassed := true
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			allPassed = runTest(t, tc) && allPassed
		})
	}

	if !allPassed {
		t.Run("find good seeds", func(t *testing.T) {
			for _, tc := range tests {
				for !runTest(t, tc) {
					tc.seed += 1
				}
			}
		})
		t.Run("good seeds", func(t *testing.T) {
			for _, tc := range tests {
				t.Logf("%d => %q", tc.seed, tc.name)
			}
			t.Fail()
		})
	}
}
