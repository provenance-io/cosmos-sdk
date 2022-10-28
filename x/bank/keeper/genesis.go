package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// InitGenesis initializes the bank module's state from a given genesis state.
func (k BaseKeeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, se := range genState.GetAllSendEnabled() {
		k.SetSendEnabled(ctx, se.Denom, se.Enabled)
	}

	totalSupply := sdk.Coins{}
	quarantinedFundsHolder := k.GetQuarantinedFundsHolder()
	quarantinedFundsHolderSupply := sdk.Coins{}
	genState.Balances = types.SanitizeGenesisBalances(genState.Balances)

	for _, balance := range genState.Balances {
		addr := balance.GetAddress()

		if err := k.initBalances(ctx, addr, balance.Coins); err != nil {
			panic(fmt.Errorf("error on setting balances %w", err))
		}

		if quarantinedFundsHolder.Equals(addr) {
			quarantinedFundsHolderSupply = quarantinedFundsHolderSupply.Add(balance.Coins...)
		}

		totalSupply = totalSupply.Add(balance.Coins...)
	}

	totalQuarantined := sdk.Coins{}
	for _, qf := range genState.QuarantinedFunds {
		toAddr := sdk.MustAccAddressFromBech32(qf.ToAddress)
		fromAddr := sdk.MustAccAddressFromBech32(qf.FromAddress)
		k.SetQuarantineRecord(ctx, toAddr, fromAddr, qf.AsQuarantineRecord())
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	if !totalQuarantined.Empty() {
		if quarantinedFundsHolderSupply.Empty() {
			if err := k.initBalances(ctx, quarantinedFundsHolder, totalQuarantined); err != nil {
				panic(fmt.Errorf("error setting balance for quarantined fund holder %w", err))
			}
			totalSupply = totalSupply.Add(quarantinedFundsHolderSupply...)
		} else {
			_, hasNeg := quarantinedFundsHolderSupply.SafeSub(totalQuarantined...)
			if hasNeg {
				panic(fmt.Errorf("quarantine fund holder account %s does not have enough funds %v to cover all quarantined funds %v",
					k.GetQuarantinedFundsHolder().String(), quarantinedFundsHolderSupply, totalQuarantined))
			}
		}
	}

	if !genState.Supply.Empty() && !genState.Supply.IsEqual(totalSupply) {
		panic(fmt.Errorf("genesis supply is incorrect, expected %v, got %v", genState.Supply, totalSupply))
	}

	for _, supply := range totalSupply {
		k.setSupply(ctx, supply)
	}

	for _, meta := range genState.DenomMetadata {
		k.SetDenomMetaData(ctx, meta)
	}

	for _, toAddrStr := range genState.QuarantinedAddresses {
		toAddr := sdk.MustAccAddressFromBech32(toAddrStr)
		k.SetQuarantineOptIn(ctx, toAddr)
	}

	for _, qar := range genState.QuarantineAutoResponses {
		toAddr := sdk.MustAccAddressFromBech32(qar.ToAddress)
		fromAddr := sdk.MustAccAddressFromBech32(qar.FromAddress)
		k.SetQuarantineAutoResponse(ctx, toAddr, fromAddr, qar.Response)
	}
}

// ExportGenesis returns the bank module's genesis state.
func (k BaseKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	totalSupply, _, err := k.GetPaginatedTotalSupply(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(fmt.Errorf("unable to fetch total supply %v", err))
	}

	rv := types.NewGenesisState(
		k.GetParams(ctx),
		k.GetAccountsBalances(ctx),
		totalSupply,
		k.GetAllDenomMetaData(ctx),
		k.GetAllSendEnabledEntries(ctx),
		k.GetAllQuarantinedAccounts(ctx),
		k.GetAllQuarantineAutoResponseEntries(ctx),
		k.GetAllQuarantinedFunds(ctx),
	)
	return rv
}
