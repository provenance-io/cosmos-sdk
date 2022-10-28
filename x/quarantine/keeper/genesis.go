package keeper

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

func (k Keeper) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState quarantine.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)

	for _, toAddrStr := range genesisState.QuarantinedAddresses {
		toAddr := sdk.MustAccAddressFromBech32(toAddrStr)
		k.SetQuarantineOptIn(ctx, toAddr)
	}

	for _, qar := range genesisState.QuarantineAutoResponses {
		toAddr := sdk.MustAccAddressFromBech32(qar.ToAddress)
		fromAddr := sdk.MustAccAddressFromBech32(qar.FromAddress)
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, qar.Response)
	}

	totalQuarantined := sdk.Coins{}
	for _, qf := range genesisState.QuarantinedFunds {
		toAddr := sdk.MustAccAddressFromBech32(qf.ToAddress)
		fromAddr := sdk.MustAccAddressFromBech32(qf.FromAddress)
		k.SetQuarantineRecord(ctx, toAddr, fromAddr, qf.AsQuarantineRecord())
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	// TODO[1046]: Use the bank keeper to make sure the fund-holder has enough funds.
}

func (k Keeper) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) *quarantine.GenesisState {
	qAddrs := k.GetAllQuarantinedAccounts(ctx)
	autoResps := k.GetAllQuarantineAutoResponseEntries(ctx)
	qFunds := k.GetAllQuarantinedFunds(ctx)

	return quarantine.NewGenesisState(qAddrs, autoResps, qFunds)
}
