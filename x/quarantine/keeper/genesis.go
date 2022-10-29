package keeper

import (
	"encoding/json"
	"fmt"

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

	if !totalQuarantined.IsZero() {
		qFundHolderBalance := k.bankKeeper.GetAllBalances(ctx, k.quarantinedFundsHolder)
		if _, hasNeg := qFundHolderBalance.SafeSub(totalQuarantined...); hasNeg {
			panic(fmt.Errorf("quarantine fund holder account %q does not have enough funds %q to cover quarantined funds %q",
				k.quarantinedFundsHolder.String(), qFundHolderBalance.String(), totalQuarantined.String()))
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) *quarantine.GenesisState {
	qAddrs := k.GetAllQuarantinedAccounts(ctx)
	autoResps := k.GetAllQuarantineAutoResponseEntries(ctx)
	qFunds := k.GetAllQuarantinedFunds(ctx)

	return quarantine.NewGenesisState(qAddrs, autoResps, qFunds)
}
