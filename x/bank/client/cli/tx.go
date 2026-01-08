package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/address"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
)

var FlagSplit = "split"

// NewTxCmd returns a root CLI command handler for all x/bank transaction commands.
func NewTxCmd(ac address.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Bank transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewSendTxCmd(ac),
		NewMultiSendTxCmd(ac),
		GetCmdSetDenomMetadata(),
	)

	return txCmd
}

// NewSendTxCmd returns a CLI command handler for creating a MsgSend transaction.
func NewSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send [from_key_or_address] [to_address] [amount]",
		Short: "Send funds from one account to another.",
		Long: `Send funds from one account to another.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address].
When using '--dry-run' a key name cannot be used, only a bech32 address.
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			toAddr, err := ac.StringToBytes(args[1])
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			if len(coins) == 0 {
				return fmt.Errorf("invalid coins")
			}

			msg := types.NewMsgSend(clientCtx.GetFromAddress(), toAddr, coins)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewMultiSendTxCmd returns a CLI command handler for creating a MsgMultiSend transaction.
// For a better UX this command is limited to send funds from one account to two or more accounts.
func NewMultiSendTxCmd(ac address.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi-send [from_key_or_address] [to_address_1 to_address_2 ...] [amount]",
		Short: "Send funds from one account to two or more accounts.",
		Long: `Send funds from one account to two or more accounts.
By default, sends the [amount] to each address of the list.
Using the '--split' flag, the [amount] is split equally between the addresses.
Note, the '--from' flag is ignored as it is implied from [from_key_or_address] and 
separate addresses with space.
When using '--dry-run' a key name cannot be used, only a bech32 address.`,
		Example: fmt.Sprintf("%s tx bank multi-send cosmos1... cosmos1... cosmos1... cosmos1... 10stake", version.AppName),
		Args:    cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[len(args)-1])
			if err != nil {
				return err
			}

			if coins.IsZero() {
				return fmt.Errorf("must send positive amount")
			}

			split, err := cmd.Flags().GetBool(FlagSplit)
			if err != nil {
				return err
			}

			totalAddrs := sdkmath.NewInt(int64(len(args) - 2))
			// coins to be received by the addresses
			sendCoins := coins
			if split {
				sendCoins = coins.QuoInt(totalAddrs)
			}

			var output []types.Output
			for _, arg := range args[1 : len(args)-1] {
				toAddr, err := ac.StringToBytes(arg)
				if err != nil {
					return err
				}

				output = append(output, types.NewOutput(toAddr, sendCoins))
			}

			// amount to be send from the from address
			var amount sdk.Coins
			if split {
				// user input: 1000stake to send to 3 addresses
				// actual: 333stake to each address (=> 999stake actually sent)
				amount = sendCoins.MulInt(totalAddrs)
			} else {
				amount = coins.MulInt(totalAddrs)
			}

			msg := types.NewMsgMultiSend(types.NewInput(clientCtx.FromAddress, amount), output)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().Bool(FlagSplit, false, "Send the equally split token amount to each address")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSetDenomMetadata returns a CLI command handler for creating a governance
// proposal to update bank denom metadata.
//
// This command constructs a MsgUpdateDenomMetadata wrapped in a governance
// proposal, allowing on-chain updates to denomination metadata such as name,
// symbol, display unit, and exponent.
//
// The command requires exactly six arguments:
//
//	<denom> <name> <symbol> <denom-description> <display> <exponent>
//
// Proposal title and summary must be provided via governance flags.
func GetCmdSetDenomMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-denom-metadata <denom> <name> <symbol> <denom-description> <display> <exponent>",
		Aliases: []string{"sdm"},
		Args:    cobra.ExactArgs(6),
		Short:   "Create a governance proposal to set denom metadata",
		Long: strings.TrimSpace(`Create a governance proposal to set denomination metadata.
This creates a gov proposal with title and description that wraps the denom metadata update.`),
		Example: fmt.Sprintf(`$ %[1]s tx bank set-denom-metadata mycoin "My Coin" "MYC" "My coin description" "myc" 6 \
  --title="Update MyCoin Metadata" \
  --description="Proposal to update metadata for mycoin" \
  --from mykey`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]
			name := args[1]
			symbol := args[2]
			denomDescription := args[3]
			display := args[4]

			exponent, err := strconv.ParseUint(args[5], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid exponent %q: %w", args[5], err)
			}

			title, _ := cmd.Flags().GetString(govcli.FlagTitle)
			description, _ := cmd.Flags().GetString(govcli.FlagSummary)

			if strings.TrimSpace(title) == "" {
				return fmt.Errorf(`required flag(s) "title" not set`)
			}
			if strings.TrimSpace(description) == "" {
				return fmt.Errorf(`required flag(s) "summary" not set`)
			}

			metadata := types.Metadata{
				Description: denomDescription,
				DenomUnits: []*types.DenomUnit{
					{
						Denom:    denom,
						Exponent: 0,
						Aliases:  []string{},
					},
					{
						Denom:    display,
						Exponent: uint32(exponent),
						Aliases:  []string{},
					},
				},
				Base:    denom,
				Display: display,
				Name:    name,
				Symbol:  symbol,
				URI:     "",
				URIHash: "",
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := &types.MsgUpdateDenomMetadata{
				FromAddress: fromAddress,
				Title:       title,
				Description: description,
				Metadata:    metadata,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)

	return cmd
}
