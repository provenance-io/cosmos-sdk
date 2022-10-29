package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func QueryCmd(name string) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Querying commands for the quarantine module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// TODO[1046]: Create the query command stuff and queryCmd.AddCommand

	return queryCmd
}

// TODO[1046]: Command for QuarantinedFunds
// TODO[1046]: Command for IsQuarantined
// TODO[1046]: Command for AutoResponses
