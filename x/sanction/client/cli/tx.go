package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// exampleTxCmdBase is the base command that gets a user to one of the tx commands in here.
var exampleTxCmdBase = fmt.Sprintf("%s tx %s", version.AppName, sanction.ModuleName)

// TxCmd returns the command with sub-commands for specific quarantine module Tx interaction.
func TxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        sanction.ModuleName,
		Short:                      "Quarantine transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
	// TODO[1046]: Add tx commmands.
	)

	return txCmd
}
