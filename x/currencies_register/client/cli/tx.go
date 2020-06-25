package cli

import (
	"bufio"
	"fmt"
	"strconv"

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
		Use:     "add-currency-proposal [denom] [decimals] [totalSupply] [path] [owner] [isToken] [flags]",
		Args:    cobra.ExactArgs(6),
		Short:   "Submit a add currency proposal",
		Example: "add-currency-proposal dfi 18 100000000000000000000 01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d my_address false --deposit 100dfi --fees 1dfi",
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

			var owner []byte
			if addr, err := sdk.AccAddressFromBech32(args[4]); err == nil {
				owner = addr.Bytes()
			} else if addr, err := sdk.AccAddressFromHex(args[4]); err == nil {
				owner = addr.Bytes()
			} else {
				return fmt.Errorf("%s argument %q parse error: Bech32 and HEX strings parsing failed", "owner", args[4])
			}

			isToken, err := strconv.ParseBool(args[5])
			if err != nil {
				return fmt.Errorf("%s argument %q parse error: %w", "isToken", args[5], err)
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
			content := types.NewAddCurrencyProposal(denom, decimals, isToken, owner, path, totalSupply)
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
