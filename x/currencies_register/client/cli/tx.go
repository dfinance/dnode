package cli

import (
	"bufio"
	"fmt"

	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govCli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// Send governance add currency proposal.
func AddCurrencyProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-currency-proposal [denom] [decimals] [totalSupply] [path] [flags]",
		Args:    cobra.ExactArgs(4),
		Short:   "Submit currency add proposal, creating non-token currency",
		Example: "add-currency-proposal dfi 18 100000000000000000000 01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d --deposit 100dfi --fees 1dfi",
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := cliBldrCtx.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// parse inputs
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)
			fromAddress := cliCtx.FromAddress
			if err := accGetter.EnsureExists(fromAddress); err != nil {
				return fmt.Errorf("%s flag: %v", flags.FlagFrom, err)
			}

			denom, decimals, totalSupply, path, err := parseCurrencyArgs(args[0], args[1], args[2], args[3])
			if err != nil {
				return err
			}

			depositStr, err := cmd.Flags().GetString(govCli.FlagDeposit)
			if err != nil {
				return fmt.Errorf("%s flag: %w", govCli.FlagDeposit, err)
			}
			deposit, err := sdk.ParseCoins(depositStr)
			if err != nil {
				return fmt.Errorf("%s flag %q: parsing: %w", govCli.FlagDeposit, depositStr, err)
			}

			// prepare and send message
			content := types.NewAddCurrencyProposal(denom, decimals, path, totalSupply)
			if err := content.ValidateBasic(); err != nil {
				return err
			}

			msg := gov.NewMsgSubmitProposal(content, deposit, fromAddress)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(govCli.FlagDeposit, "", "deposit of proposal")

	return cmd
}
