package cli

import (
	"encoding/hex"
	"fmt"
	"github.com/dfinance/dnode/helpers"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCmdAssetCodeHex returns converted asset code in the hex format.
func GetCmdAssetCodeHex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset-code-hex [assetCode]",
		Short: "get asset code in hex",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			res := hex.EncodeToString([]byte(assetCode.String()))
			fmt.Printf("Asset code in hex: %s\n", res)
			return nil
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"assetCode [string] asset code symbol",
	})

	return cmd
}

// GetCmdCurrentPrice returns the current price of an asset.
func GetCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [assetCode]",
		Short: "get the current price of an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, types.QueryPrice, assetCode.String()), nil)
			if err != nil {
				return err
			}

			var out types.CurrentPrice
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}

	helpers.BuildCmdHelp(cmd, []string{
		"assetCode [string] asset code symbol",
	})

	return cmd
}

// GetCmdRawPrices returns the raw price of an asset.
func GetCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawprices [assetCode] [blockHeight]",
		Short: "get the raw oracle prices for an asset",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			blockHeight, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("%s argument %q not a number: %v", "blockHeight", args[1], err)
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%d", queryRoute, types.QueryRawPrices, assetCode, blockHeight), nil)
			if err != nil {
				fmt.Printf("could not get raw prices for - %s/%d \n", assetCode, blockHeight)
				return nil
			}

			var out types.QueryRawPricesResp
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdAssets returns the list of assets of an oracle.
func GetCmdAssets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "assets",
		Short: "get the list of assets in the oracle",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAssets), nil)
			if err != nil {
				fmt.Printf("could not get assets %v", err)
				return nil
			}

			var out types.Assets
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
