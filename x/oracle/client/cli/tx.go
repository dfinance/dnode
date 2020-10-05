package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCmdPostPrice returns tx command for posting price for a particular asset.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "postprice [assetCode] [askPrice] [bidPrice] [receivedAt]",
		Short:   "Post the latest price for a particular asset",
		Example: "postprice eth_usdt 100 95 1594732456",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			askPrice, err := helpers.ParseSdkIntParam("askPrice", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			bidPrice, err := helpers.ParseSdkIntParam("bidPrice", args[2], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			receivedAt, err := helpers.ParseUnixTimestamp("receivedAt", args[3], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgPostPrice(fromAddr, assetCode, askPrice, bidPrice, receivedAt)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"askPrice [int]",
		"bidPrice [int]",
		"price received at UNIX timestamp in seconds [int]",
	})

	return cmd
}

// GetCmdAddOracle returns tx command for adding new oracle for a particular asset.
func GetCmdAddOracle(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-oracle [assetCode] [oracleAddress]",
		Short:   "Add a new oracle for a particular asset",
		Example: "add-oracle eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracleAddr, err := helpers.ParseSdkAddressParam("oracleAddress", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgAddOracle(fromAddr, assetCode, oracleAddr)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"oracle addresses",
	})

	return cmd
}

// GetCmdSetOracles returns tx command which sets oracles for a particular asset.
func GetCmdSetOracles(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-oracles [assetCode] [oracleAddresses]",
		Short:   "Sets a list of oracles for a particular asset",
		Example: "set-oracles eth_usdt wallet10ff6y8gm2re6awfwz5dvesar8jq02tx7vcvuxn,wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := parseOraclesArg("oracleAddresses", args[1])
			if err != nil {
				return err
			}

			// prepare and send message
			msg := types.NewMsgSetOracles(fromAddr, assetCode, oracles)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"comma separated list of oracle addresses",
	})

	return cmd
}

// GetCmdAddAsset returns tx command for adding a new asset for a list of oracles.
func GetCmdAddAsset(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-asset [assetCode] [oracleAddresses]",
		Short:   "Add a new asset for a list of oracles",
		Example: "dncli oracle add-asset eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := parseOraclesArg("oracleAddresses", args[1])
			if err != nil {
				return err
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracleAddresses", args[1])
			}

			// prepare and send message
			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgAddAsset(fromAddr, asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"comma separated list of oracle addresses",
	})

	return cmd
}

// GetCmdSetAsset returns tx command which sets (updates) an existing asset for a list of oracles.
func GetCmdSetAsset(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "set-asset [assetCode] [oracleAddresses]",
		Short:   "Set the existing asset for a list of oracles",
		Example: "set-asset eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx, txBuilder := helpers.GetTxCmdCtx(cdc, cmd.InOrStdin())

			// parse inputs
			fromAddr, err := helpers.ParseFromFlag(cliCtx)
			if err != nil {
				return err
			}

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			oracles, err := parseOraclesArg("oracleAddresses", args[1])
			if err != nil {
				return err
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracleAddresses", args[1])
			}

			// prepare and send message
			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgSetAsset(fromAddr, asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBuilder, []sdk.Msg{msg})
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"comma separated list of oracle addresses",
	})

	return cmd
}
