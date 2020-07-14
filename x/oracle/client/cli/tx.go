package cli

import (
	"bufio"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/spf13/cobra"
	tmtime "github.com/tendermint/tendermint/types/time"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCmdPostPrice cli command for posting prices.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "postprice [fromKeyOrAddress] [assetCode] [price] [receivedAt]",
		Short: "post the latest price for a particular asset",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[1])
			if err := assetCode.Validate(); err != nil {
				return err
			}

			rawPrice := args[2]
			price, ok := sdk.NewIntFromString(rawPrice)
			if !ok {
				return fmt.Errorf("%s argument %q: wrong value for price", "price", args[2])
			}
			receivedAtInt, ok := sdk.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("%s argument %q: wrong value for time", "receivedAt", args[3])
			}
			receivedAt := tmtime.Canonical(time.Unix(receivedAtInt.Int64(), 0))
			msg := types.NewMsgPostPrice(cliCtx.GetFromAddress(), assetCode, price, receivedAt)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdAddOracle cli command for create new oracle.
func GetCmdAddOracle(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-oracle [nomineeKey] [assetCode] [oracle_address]",
		Example: "dncli oracle add-oracle wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a new oracle",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[1])
			if err := assetCode.Validate(); err != nil {
				return fmt.Errorf("%s argument %q: %w", "assetCode", args[1], err)
			}

			oracleAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracle_address", args[2], err)
			}

			msg := types.NewMsgAddOracle(cliCtx.GetFromAddress(), assetCode, oracleAddr)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetOracles cli command for set a list of oracles for a denom.
func GetCmdSetOracles(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-oracles [nomineeKey] [assetCode] [oracle_addresses]",
		Example: "dncli oracle set-oracles wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet10ff6y8gm2re6awfwz5dvesar8jq02tx7vcvuxn,wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Sets a list of oracles for a denom",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[1])
			if err := assetCode.Validate(); err != nil {
				return fmt.Errorf("%s argument %q: %w", "assetCode", args[1], err)
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracle_addresses", args[2], err)
			}

			msg := types.NewMsgSetOracles(cliCtx.GetFromAddress(), assetCode, oracles)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdAddAsset cli command for create a new asset.
func GetCmdAddAsset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-asset [nomineeKey] [assetCode] [oracles]",
		Example: "dncli oracle add-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a new asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[1])
			if err := assetCode.Validate(); err != nil {
				return fmt.Errorf("%s argument %q: %w", "assetCode", args[1], err)
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "oracles", args[2], err)
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracles", args[2])
			}

			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgAddAsset(cliCtx.GetFromAddress(), asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdSetAsset cli command for set asset.
func GetCmdSetAsset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-asset [nomineeKey] [assetCode] [oracles]",
		Example: "dncli oracle set-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a set asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[1])
			if err := assetCode.Validate(); err != nil {
				return fmt.Errorf("%s argument %q: %w", "assetCode", args[1], err)
			}

			oracles, err := types.ParseOracles(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %v", "oracles", args[2], err)
			}
			if len(oracles) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "oracles", args[2])
			}

			asset := types.NewAsset(assetCode, oracles, true)
			if err := asset.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgSetAsset(cliCtx.GetFromAddress(), asset)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
