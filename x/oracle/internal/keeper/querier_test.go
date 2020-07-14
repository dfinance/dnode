// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func TestOracleKeeper_QueryCurrentPrice(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.keeper, input.ctx

	// get current price ok
	{
		_, err := queryCurrentPrice(ctx, []string{input.stdAssetCode.String()}, abci.RequestQuery{}, keeper)
		require.NoError(t, err)
	}

	// wrong asset code
	{
		_, err := queryCurrentPrice(ctx, []string{"wrong_asset"}, abci.RequestQuery{}, keeper)
		require.Error(t, err)
	}

	// empty params
	{
		defer func() {
			r := recover()
			require.NotNil(t, r)
		}()

		_, _ = queryCurrentPrice(ctx, []string{}, abci.RequestQuery{}, keeper)
	}
}

func TestOracleKeeper_QueryRawPrices(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.keeper, input.ctx

	// get current price ok
	{
		_, err := queryRawPrices(ctx, []string{input.stdAssetCode.String(), "1"}, abci.RequestQuery{}, keeper)
		require.NoError(t, err)
	}

	// wrong block
	{
		_, err := queryRawPrices(ctx, []string{input.stdAssetCode.String(), "block"}, abci.RequestQuery{}, keeper)
		require.Error(t, err)
	}

	// wrong asset code
	{
		_, err := queryRawPrices(ctx, []string{"wrong_asset", "1"}, abci.RequestQuery{}, keeper)
		require.Error(t, err)
	}

	// empty params
	{
		defer func() {
			r := recover()
			require.NotNil(t, r)
		}()

		_, _ = queryRawPrices(ctx, []string{}, abci.RequestQuery{}, keeper)
	}

	// empty block height
	{
		defer func() {
			r := recover()
			require.NotNil(t, r)
		}()
		_, _ = queryRawPrices(ctx, []string{input.stdAssetCode.String()}, abci.RequestQuery{}, keeper)
	}
}

func TestOracleKeeper_QueryAssets(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.keeper, input.ctx, input.cdc

	// ok
	{
		resp, err := queryAssets(ctx, abci.RequestQuery{}, keeper)
		require.NoError(t, err)
		var assets types.Assets
		err = cdc.UnmarshalJSON(resp, &assets)

		require.NoError(t, err)
		require.Len(t, assets, len(input.stdAssets))
		require.Equal(t, assets[0].AssetCode, input.stdAssets[0].AssetCode)
	}
}
