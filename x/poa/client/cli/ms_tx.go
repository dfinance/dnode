// Operations with validators via multisignature calls by CLI.
package cli

import (
	"os"

	msMsg "github.com/WingsDao/wings-blockchain/x/multisig/msgs"
	"github.com/WingsDao/wings-blockchain/x/poa/msgs"

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
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			ethAddress := args[1]
			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err != nil {
				return err
			}

			addVldrMsg := msgs.NewMsgAddValidator(validatorAddress, ethAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(addVldrMsg, args[2], cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
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
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err != nil {
				return err
			}

			msgRmvVal := msgs.NewMsgRemoveValidator(validatorAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(msgRmvVal, args[1], cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
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
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			accGetter := txBldrCtx.NewAccountRetriever(cliCtx)

			if err := accGetter.EnsureExists(cliCtx.FromAddress); err != nil {
				return err
			}

			oldValidator, err := sdk.AccAddressFromBech32(args[0])

			if err != nil {
				return err
			}

			newValidator, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			ethAddress := args[2]

			msgReplVal := msgs.NewMsgReplaceValidator(oldValidator, newValidator, ethAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(msgReplVal, args[3], cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg})
		},
	}
}
