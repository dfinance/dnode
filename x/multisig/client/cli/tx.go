// Implements TX queries for modules.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"

	msMsg "github.com/dfinance/dnode/x/multisig/msgs"
)

// Post confirmation for multisig call via CLI.
func PostConfirmCall(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "confirm-call [callId]",
		Short: "confirm call by multisig",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			callId, err := strconv.ParseUint(args[0], 10, 8)
			if err != nil {
				return fmt.Errorf("%s argument %q: parsing uint: %w", "callId", args[0], err)
			}

			msg := msMsg.NewMsgConfirmCall(callId, cliCtx.GetFromAddress())

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// Post revoke confirmation for multisig call via CLI.
func PostRevokeConfirm(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke-confirm [callId]",
		Short: "revoke confirmation from call by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			callId, err := strconv.ParseUint(args[0], 10, 8)
			if err != nil {
				return fmt.Errorf("%s argument %q: parsing uint: %w", "callId", args[0], err)
			}

			msg := msMsg.NewMsgRevokeConfirm(callId, cliCtx.GetFromAddress())

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
