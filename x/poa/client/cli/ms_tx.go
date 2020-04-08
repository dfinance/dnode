// Operations with validators via multisignature calls by CLI.
package cli

import (
	"bufio"
	"fmt"
	"os"

	msMsg "github.com/dfinance/dnode/x/multisig/msgs"
	"github.com/dfinance/dnode/x/poa/msgs"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

// Add new validator via multisignature.
func PostMsAddValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-add-validator [address] [ethAddress] [uniqueID]",
		Short: "adding new validator to validator list by multisig",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			ethAddress := args[1]
			validatorAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "address", args[0], err)
			}

			addVldrMsg := msgs.NewMsgAddValidator(validatorAddress, ethAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(addVldrMsg, args[2], cliCtx.GetFromAddress())

			if err := msMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg})
		},
	}
}

// Remove validator via multisignature.
func PostMsRemoveValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-remove-validator [address] [uniqueID]",
		Short: "remove poa validator by multisig",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			validatorAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "address", args[0], err)
			}

			msgRmvVal := msgs.NewMsgRemoveValidator(validatorAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(msgRmvVal, args[1], cliCtx.GetFromAddress())

			if err := msMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg})
		},
	}
}

// Replace validator via multisignature.
func PostMsReplaceValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-replace-validator [oldValidator] [newValidator] [ethAddress] [uniqueID]",
		Short: "replace poa validator by multisignature",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := txBldrCtx.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return fmt.Errorf("fromAddress: %w", err)
			}

			oldValidator, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oldValidator", args[0], err)
			}

			newValidator, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "newValidator", args[1], err)
			}

			ethAddress := args[2]

			msgReplVal := msgs.NewMsgReplaceValidator(oldValidator, newValidator, ethAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(msgReplVal, args[3], cliCtx.GetFromAddress())

			if err := msMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg})
		},
	}
}
