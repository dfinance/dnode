package cli

import (
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	msClient "github.com/dfinance/dnode/x/multisig/client"
	"github.com/dfinance/dnode/x/poa/internal/types"
)

// PostMsAddValidator returns tx command which post a new multisig add validator request.
func PostMsAddValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ms-add-validator [uniqueID] [address] [ethAddress]",
		Short:   "Add a new PoA validator to the validator list via multisignature",
		Example: "ms-add-validator add1 {validatorAccount} 0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d --from {account}",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			sdkAddr, err := helpers.ParseSdkAddressParam("address", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			ethAddr, err := helpers.ParseEthereumAddressParam("ethAddress", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send multisig message
			msg := types.NewMsgAddValidator(sdkAddr, ethAddr, fromAddr)
			callMsg := msClient.NewMsgSubmitCall(msg, args[0], fromAddr)
			if err := callMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{callMsg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"unique multi signature call ID",
		"validator SDK address",
		"validator Ethereum address",
	})

	return cmd
}

// PostMsRemoveValidator returns tx command which post a new multisig remove validator request.
func PostMsRemoveValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ms-remove-validator [uniqueID] [address]",
		Short:   "Remove a PoA validator from the validator list via multisignature",
		Example: "ms-remove-validator remove1 {validatorAccount} --from {account}",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			sdkAddr, err := helpers.ParseSdkAddressParam("address", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send multisig message
			msg := types.NewMsgRemoveValidator(sdkAddr, fromAddr)
			msMsg := msClient.NewMsgSubmitCall(msg, args[0], fromAddr)
			if err := msMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msMsg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"unique multi signature call ID",
		"validator SDK address",
	})

	return cmd
}

// PostMsReplaceValidator returns tx command which post a new multisig replace validator request.
func PostMsReplaceValidator(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ms-replace-validator [uniqueID] [oldValidator] [newValidator] [ethAddress]",
		Short: "Replace an old PoA validator with the new one via multisignature",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			sdkAddrOld, err := helpers.ParseSdkAddressParam("oldValidator", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			sdkAddrNew, err := helpers.ParseSdkAddressParam("newValidator", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			ethAddr, err := helpers.ParseEthereumAddressParam("ethAddress", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send multisig message
			msg := types.NewMsgReplaceValidator(sdkAddrOld, sdkAddrNew, ethAddr, fromAddr)
			msMsg := msClient.NewMsgSubmitCall(msg, args[0], fromAddr)
			if err := msMsg.ValidateBasic(); err != nil {
				return err
			}

			cliCtx.WithOutput(os.Stdout)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msMsg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"unique multi signature call ID",
		"old validator SDK address to replace",
		"new validator SDK address to replace",
		"new validator Ethereum address",
	})

	return cmd
}
