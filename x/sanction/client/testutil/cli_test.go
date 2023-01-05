//go:build norace
// +build norace

package testutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 2
	cfg.TimeoutCommit = 1 * time.Second
	sanctionedAddr1 := sdk.AccAddress("1_sanctioned_address_")
	sanctionedAddr2 := sdk.AccAddress("2_sanctioned_address_")
	tempSanctAddr := sdk.AccAddress("temp_sanctioned_addr")
	tempUnsanctAddr := sdk.AccAddress("temp_unsanctioned___")
	sanctionGenBz := cfg.GenesisState[sanction.ModuleName]
	var sanctionGen sanction.GenesisState
	if len(sanctionGenBz) > 0 {
		cfg.Codec.MustUnmarshalJSON(sanctionGenBz, &sanctionGen)
	}
	sanctionGen.SanctionedAddresses = append(sanctionGen.SanctionedAddresses,
		sanctionedAddr1.String(),
		sanctionedAddr2.String(),
	)
	sanctionGen.TemporaryEntries = append(sanctionGen.TemporaryEntries,
		&sanction.TemporaryEntry{
			Address:    tempSanctAddr.String(),
			ProposalId: 1,
			Status:     sanction.TEMP_STATUS_SANCTIONED,
		},
		&sanction.TemporaryEntry{
			Address:    tempUnsanctAddr.String(),
			ProposalId: 1,
			Status:     sanction.TEMP_STATUS_UNSANCTIONED,
		},
	)
	sanctionGen.Params = &sanction.Params{
		ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 52)),
		ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin(cfg.BondDenom, 133)),
	}
	cfg.GenesisState[sanction.ModuleName] = cfg.Codec.MustMarshalJSON(&sanctionGen)
	suite.Run(t, NewIntegrationTestSuite(cfg, &sanctionGen))
}
