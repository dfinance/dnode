package cli


import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/poa/msgs"
	msMsg "wings-blockchain/x/multisig/msgs"
)

func PostMsAddValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-add-validator [address] [ethAddress]",
		Short: "adding new validator to validator list by multisig",
		Args:  cobra.ExactArgs(2),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			ethAddress := args[1]
			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err  != nil {
				return err
			}

			addVldrMsg := msgs.NewMsgAddValidator(validatorAddress, ethAddress, cliCtx.GetFromAddress())
			msMsg := msMsg.NewMsgSubmitCall(addVldrMsg, cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg}, false)
		},
	}
}

func PostMsRemoveValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-remove-validator [address]",
		Short: "remove poa validator by multisig",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err != nil {
				return err
			}

			msgRmvVal := msgs.NewMsgRemoveValidator(validatorAddress, cliCtx.GetFromAddress())
			msMsg     := msMsg.NewMsgSubmitCall(msgRmvVal, cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg}, false)
		},
	}
}


func PostMsReplaceValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "ms-replace-validator [oldValidator] [newValidator] [ethAddress]",
		Short: "replace poa validator by multisignature",
		Args:  cobra.ExactArgs(3),
		RunE:  func(cmd *cobra.Command,  args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
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
			msMsg 	   := msMsg.NewMsgSubmitCall(msgReplVal, cliCtx.GetFromAddress())

			err = msMsg.ValidateBasic()

			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msMsg}, false)
		},
	}
}

