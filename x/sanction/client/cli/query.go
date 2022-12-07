package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/sanction"
)

// exampleQueryCmdBase is the base command that gets a user to one of the query commands in here.
var exampleQueryCmdBase = fmt.Sprintf("%s query %s", version.AppName, sanction.ModuleName)

// QueryCmd returns the command with sub-commands for specific sanction module queries.
func QueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        sanction.ModuleName,
		Short:                      "Querying commands for the quarantine module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		// TODO[1046]: Add the various query commands.
	)

	return queryCmd
}
