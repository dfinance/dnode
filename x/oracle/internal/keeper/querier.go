package keeper

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// price Takes an [assetcode] and returns CurrentPrice for that asset
// oracle Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the oracle system

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryCurrentPrice:
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

func queryCurrentPrice(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	assetCode := path[0]
	if _, found := keeper.GetAsset(ctx, assetCode); !found {
		return []byte{}, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "asset not found")
	}
	currentPrice := keeper.GetCurrentPrice(ctx, assetCode)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, currentPrice)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryRawPrices(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err error) {
	assetCode := path[0]
	if _, found := keeper.GetAsset(ctx, assetCode); !found {
		return []byte{}, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "asset not found")
	}

	blockHeight, blockErr := strconv.ParseInt(path[1], 10, 64)
	if blockErr != nil {
		return []byte{}, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "invalid blockSize: %v", blockErr)
	}

	priceList := keeper.GetRawPrices(ctx, assetCode, blockHeight)
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, priceList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

func queryAssets(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, error) {
	assets := keeper.GetAssetParams(ctx)
	bz := codec.MustMarshalJSONIndent(keeper.cdc, &assets)

	return bz, nil
}
