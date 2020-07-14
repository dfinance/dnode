package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// Convert asset code in hex.
func GetCmdAssetCodeHex() *cobra.Command {
	return &cobra.Command{
		Use:   "asset-code-hex [assetCode]",
		Short: "get asset code in hex",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			assetCode := dnTypes.AssetCode(args[0])
			if err := assetCode.Validate(); err != nil {
				return err
			}

			res := hex.EncodeToString([]byte(assetCode.String()))
			fmt.Printf("Asset code in hex: %s\n", res)
			return nil
		},
	}
}

// GetCmdCurrentPrice queries the current price of an asset.
func GetCmdCurrentPrice(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "price [assetCode]",
		Short: "get the current price of an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			assetCode := dnTypes.AssetCode(args[0])
			if err := assetCode.Validate(); err != nil {
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
}

// GetCmdRawPrices queries the current price of an asset.
func GetCmdRawPrices(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rawprices [assetCode] [blockHeight]",
		Short: "get the raw oracle prices for an asset",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			assetCode := args[0]

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

// GetCmdAssets queries list of assets in the oracle.
func GetCmdAssets(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "assets",
		Short: "get the assets in the oracle",
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
