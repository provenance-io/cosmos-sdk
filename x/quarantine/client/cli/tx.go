package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
)

func TxCmd(name string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Quarantine transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// TODO[1046]: Create the tx command stuff and txCmd.AddCommand

	return txCmd
}

// TODO[1046]: Command for OptIn
// TODO[1046]: Command for OptOut
// TODO[1046]: Command for Accept
// TODO[1046]: Command for Decline
// TODO[1046]: Command for UpdateAutoResponses
