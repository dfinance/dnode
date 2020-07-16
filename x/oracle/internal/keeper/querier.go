package keeper

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryPrice:
			return queryCurrentPrice(ctx, path[1:], req, keeper)
		case types.QueryRawPrices:
			return queryRawPrices(ctx, path[1:], req, keeper)
		case types.QueryAssets:
			return queryAssets(ctx, req, keeper)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown oracle query endpoint")
		}
	}
}

// queryCurrentPrice handles currentPrice query. Takes an [assetCode] and returns CurrentPrice for that asset.
func queryCurrentPrice(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	assetCode := dnTypes.AssetCode(path[0])
	if _, found := keeper.GetAsset(ctx, assetCode); !found {
		return []byte{}, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "asset not found")
	}
	currentPrice := keeper.GetCurrentPrice(ctx, assetCode)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, currentPrice)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "currentPrice marshal: %v", err)
	}

	return bz, nil
}

// queryRawPrices handles rawPrice query. Takes an [assetCode] and [blockHeight], then returns the raw []PostedPrice for that asset.
func queryRawPrices(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	assetCode := dnTypes.AssetCode(path[0])
	if _, found := keeper.GetAsset(ctx, assetCode); !found {
		return []byte{}, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "asset not found")
	}

	blockHeight, blockErr := strconv.ParseInt(path[1], 10, 64)
	if blockErr != nil {
		return []byte{}, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "invalid blockSize: %v", blockErr)
	}

	priceList := keeper.GetRawPrices(ctx, assetCode, blockHeight)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, priceList)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "currentPrice marshal: %v", err)
	}

	return bz, nil
}

// queryAssets handles assets query, returns []Assets in the oracle system.
func queryAssets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	assets := keeper.GetAssetParams(ctx)
	bz := codec.MustMarshalJSONIndent(keeper.cdc, &assets)

	return bz, nil
}
