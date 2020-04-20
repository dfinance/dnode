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

	"github.com/dfinance/dnode/x/pricefeed/internal/types"
)

// GetCmdPostPrice cli command for posting prices.
func GetCmdPostPrice(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "postprice [from_key_or_address] [assetCode] [price] [receivedAt]",
		Short: "post the latest price for a particular asset",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

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
			msg := types.NewMsgPostPrice(cliCtx.GetFromAddress(), args[1], price, receivedAt)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdAddPriceFeed(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-pricefeed [nominee_key] [denom] [pricefeed_address]",
		Example: "dncli pricefeed add-pricefeed wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a new price feed",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			pricefeedAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "pricefeed_address", args[2], err)
			}

			msg := types.NewMsgAddPriceFeed(cliCtx.GetFromAddress(), args[1], pricefeedAddr)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSetOracles(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-pricefeeds [nominee_key] [denom] [pricefeed_addresses]",
		Example: "dncli pricefeed set-pricefeeds wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet10ff6y8gm2re6awfwz5dvesar8jq02tx7vcvuxn,wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Sets a list of price feeds for a denom",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			pricefeed, err := types.ParsePricefeeds(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "pricefeed_addresses", args[2], err)
			}

			msg := types.NewMsgSetOracles(cliCtx.GetFromAddress(), args[1], pricefeed)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdAddAsset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "add-asset [nominee_key] [denom] [pricefeeds]",
		Example: "dncli pricefeed add-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a new asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			denom := args[1]
			if len(denom) == 0 {
				return fmt.Errorf("%s argument %q: empty", "denom", args[1])
			}

			pricefeed, err := types.ParsePricefeeds(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %w", "pricefeed", args[2], err)
			}
			if len(pricefeed) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "pricefeed", args[2])
			}

			token := types.NewAsset(denom, pricefeed, true)
			if err := token.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgAddAsset(cliCtx.GetFromAddress(), denom, token)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdSetAsset(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:     "set-asset [nominee_key] [denom] [pricefeeds]",
		Example: "dncli pricefeed set-asset wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m eth_usdt wallet1a7260dyzp487r7wghr99f6r3h2h2z4gk4d740k",
		Short:   "Create a set asset",
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInputAndFrom(inBuf, args[0]).WithCodec(cdc)

			denom := args[1]
			if len(denom) == 0 {
				return fmt.Errorf("%s argument %q: empty", "denom", args[1])
			}

			pricefeeds, err := types.ParsePricefeeds(args[2])
			if err != nil {
				return fmt.Errorf("%s argument %q: %v", "pricefeeds", args[2], err)
			}
			if len(pricefeeds) == 0 {
				return fmt.Errorf("%s argument %q: empty slice", "pricefeeds", args[2])
			}

			token := types.NewAsset(denom, pricefeeds, true)
			if err := token.ValidateBasic(); err != nil {
				return err
			}

			msg := types.NewMsgSetAsset(cliCtx.GetFromAddress(), denom, token)

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
