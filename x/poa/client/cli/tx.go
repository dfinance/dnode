package cli

import (
	"os"
	"wings-blockchain/x/poa/msgs"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	txBldrCtx "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/spf13/cobra"
)

func PostAddValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "add-validator [address] [ethAddress]",
		Short: "add new poa validator",
		Args:  cobra.ExactArgs(2),
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

			msg := msgs.NewMsgAddValidator(validatorAddress, ethAddress, cliCtx.GetFromAddress())
			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func PostReplaceValidator(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "replace-validator [oldValidator] [newValidator] [ethAddress]",
		Short: "replace poa validator",
		Args:  cobra.ExactArgs(3),
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

			msg := msgs.NewMsgReplaceValidator(oldValidator, newValidator, ethAddress, cliCtx.GetFromAddress())
			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
