package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AccountKeeper defines the account/auth functionality needed from within the quarantine module.
type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
}

// BankKeeper defines the bank functionality needed from within the quarantine module.
type BankKeeper interface {
	SetSanctionKeeper(keeper banktypes.SanctionKeeper)
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool)
}