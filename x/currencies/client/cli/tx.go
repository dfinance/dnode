package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	cliBldrCtx "github.com/cosmos/cosmos-sdk/client/context"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"strconv"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/currencies/msgs"
)

// Issue new currency command
func PostIssueCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "issue-currency [symbol] [amount] [decimals]",
		Short: "issue new currency",
		Args:  cobra.ExactArgs(3),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)

			if err != nil {
				return err
			}

			decimals, err := strconv.ParseInt(args[2], 10, 8)

			if err != nil {
				return err
			}

			msg := msgs.NewMsgIssueCurrency(args[0], amount, int8(decimals), cliCtx.GetFromAddress())
			err = msg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

// Destroy currency
func PostDestroyCurrency(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use: 	"destroy-currency [symbol] [amount]",
		Short:  "destroy issued currency",
		Args: 	cobra.ExactArgs(2),
		RunE:   func(cmd *cobra.Command, args []string) error {
			cliCtx := cliBldrCtx.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)

			if err != nil {
				return err
			}

			msg := msgs.NewMsgDestroyCurrency(args[0], amount, cliCtx.GetFromAddress())
			err = msg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
