package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

const (
	SanctionAddresses   = "sanction-addresses"
	SanctionTempEntries = "sanction-temp-entries"
	SanctionParams      = "sanction-params"
)

func RandomSanctionedAddresses(r *rand.Rand, accounts []simtypes.Account) []string {
	// each account has a 20% chance of being sanctioned
	var rv []string
	for _, acct := range accounts {
		if r.Int63n(5) == 0 {
			rv = append(rv, acct.Address.String())
		}
	}
	return rv
}

func RandomTempEntries(r *rand.Rand, accounts []simtypes.Account) []*sanction.TemporaryEntry {
	// Each account has a 10% chance of a temp sanction, 10% chance of temp unsanction, or 80% of nothing.
	var rv []*sanction.TemporaryEntry
	for _, acct := range accounts {
		switch r.Int63n(10) {
		case 0:
			rv = append(rv, &sanction.TemporaryEntry{
				Address:    acct.Address.String(),
				ProposalId: uint64(r.Int63n(1000) + 1_000_000_000),
				Status:     sanction.TEMP_STATUS_SANCTIONED,
			})
		case 1:
			rv = append(rv, &sanction.TemporaryEntry{
				Address:    acct.Address.String(),
				ProposalId: uint64(r.Int63n(1000) + 2_000_000_000),
				Status:     sanction.TEMP_STATUS_UNSANCTIONED,
			})
		}
	}
	return rv
}

func RandomParams(r *rand.Rand) *sanction.Params {
	return &sanction.Params{
		ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1_000_000_000))),
		ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1_000_000_000))),
	}
}

func RandomizedGenState(simState *module.SimulationState) {
	genState := &sanction.GenesisState{}

	// SanctionedAddresses
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionAddresses, &genState.SanctionedAddresses, simState.Rand,
		func(r *rand.Rand) { genState.SanctionedAddresses = RandomSanctionedAddresses(r, simState.Accounts) },
	)

	// TemporaryEntries
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionTempEntries, &genState.TemporaryEntries, simState.Rand,
		func(r *rand.Rand) { genState.TemporaryEntries = RandomTempEntries(r, simState.Accounts) },
	)

	// Params
	simState.AppParams.GetOrGenerate(
		simState.Cdc, SanctionParams, &genState.Params, simState.Rand,
		func(r *rand.Rand) { genState.Params = RandomParams(r) },
	)

	simState.GenState[sanction.ModuleName] = simState.Cdc.MustMarshalJSON(genState)
}
