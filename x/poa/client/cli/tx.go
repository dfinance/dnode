package cli

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/poa/msgs"
)

func PostAddValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "add-validator [address] [ethAddress]",
		Short: "add new poa validator",
		Args:  cobra.ExactArgs(2),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			ethAddress := args[1]
			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err  != nil {
				return err
			}

			msg := msgs.NewMsgAddValidator(validatorAddress, ethAddress, cliCtx.GetFromAddress())
			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func PostRemoveValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-validator [address]",
		Short: "remove poa validator",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			txBldr := txBldrCtx.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}

			validatorAddress, err := sdk.AccAddressFromBech32(args[0])

			if err != nil {
				return err
			}

			msg := msgs.NewMsgRemoveValidator(validatorAddress, cliCtx.GetFromAddress())
			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func PostReplaceValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "replace-validator [oldValidator] [newValidator] [ethAddress]",
		Short: "replace poa validator",
		Args:  cobra.ExactArgs(3),
		RunE:  func(cmd *cobra.Command,  args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
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

			msg := msgs.NewMsgReplaceValidator(oldValidator, newValidator, ethAddress, cliCtx.GetFromAddress())
			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}