package simulation

import (
	"github.com/cosmos/cosmos-sdk/types/module"
)

// RandomizedGenState generates a random GenesisState for the quarantine module.
func RandomizedGenState(simState *module.SimulationState) {
	// TODO[1046]: Create teh random sate.
	// simState.AppParams.GetOrGenerate(...)
	//
	// quarantineGenesis := quarantine.GenesisState{...}
	// simState.GenState[group.ModuleName] = simState.Cdc.MustMarshalJSON(&groupGenesis)
}
