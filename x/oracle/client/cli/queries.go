package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"github.com/dfinance/dnode/helpers"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetCmdAssetCodeHex returns query command that returns converted asset code in the hex format.
func GetCmdAssetCodeHex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset-code-hex [assetCode]",
		Short: "Get asset code in hex",
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
		"asset code symbol",
	})

	return cmd
}

// GetCmdCurrentPrice returns query command that returns current price of an asset.
func GetCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [assetCode]",
		Short: "Get the current price for the asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// query and parse the result
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
		"asset code symbol",
	})

	return cmd
}

// GetCmdRawPrices returns query command that returns raw prices for the asset.
func GetCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rawprices [assetCode] [blockHeight]",
		Short: "Get raw oracle prices for an asset for specific block height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// parse inputs
			assetCode, err := helpers.ParseAssetCodeParam("assetCode", args[0], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			blockHeight, err := helpers.ParseUint64Param("blockHeight", args[1], helpers.ParamTypeCliArg)
			if err != nil {
				return err
			}

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%d", queryRoute, types.QueryRawPrices, assetCode, blockHeight), nil)
			if err != nil {
				fmt.Printf("could not get raw prices for %s/%d\n", assetCode, blockHeight)
				return nil
			}

			var out types.QueryRawPricesResp
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
	helpers.BuildCmdHelp(cmd, []string{
		"asset code symbol",
		"block height [uint]",
	})

	return cmd
}

// GetCmdAssets returns query command that returns list of assets.
func GetCmdAssets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "assets",
		Short: "Get assets list",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query and parse the result
			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryAssets), nil)
			if err != nil {
				return err
			}

			var out types.Assets
			cdc.MustUnmarshalJSON(res, &out)

			return cliCtx.PrintOutput(out)
		},
	}
}
