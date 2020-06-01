// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

func Test_Market_StoreIO(t *testing.T) {
	input := NewTestInput(t)

	// non-existing market
	{
		id, _ := dnTypes.NewIDFromString("0")
		require.False(t, input.keeper.Has(input.ctx, id))

		_, err := input.keeper.Get(input.ctx, id)
		require.Error(t, err)

		_, err = input.keeper.GetExtended(input.ctx, id)
		require.Error(t, err)
	}

	// add market
	var marketID dnTypes.ID
	{
		market, err := input.keeper.Add(input.ctx, input.baseBtcDenom, input.quoteDenom)
		require.NoError(t, err)
		require.Equal(t, market.BaseAssetDenom, input.baseBtcDenom)
		require.Equal(t, market.QuoteAssetDenom, input.quoteDenom)
		marketID = market.ID
	}

	// get market
	{
		market, err := input.keeper.Get(input.ctx, marketID)
		require.NoError(t, err)
		require.True(t, market.ID.Equal(marketID))
		require.Equal(t, market.BaseAssetDenom, input.baseBtcDenom)
		require.Equal(t, market.QuoteAssetDenom, input.quoteDenom)
	}

	// get extended market
	{
		extMarket, err := input.keeper.GetExtended(input.ctx, marketID)
		require.NoError(t, err)

		require.True(t, extMarket.ID.Equal(marketID))
		require.Equal(t, string(extMarket.BaseCurrency.Denom), input.baseBtcDenom)
		require.Equal(t, string(extMarket.QuoteCurrency.Denom), input.quoteDenom)
		require.Equal(t, extMarket.BaseCurrency.Decimals, input.baseBtcDecimals)
		require.Equal(t, extMarket.QuoteCurrency.Decimals, input.quoteDecimals)
	}
}

func Test_Market_List(t *testing.T) {
	input := NewTestInput(t)

	// get empty list
	{
		list := input.keeper.GetList(input.ctx)
		require.Len(t, list, 0)
	}

	market1, err := input.keeper.Add(input.ctx, input.baseBtcDenom, input.quoteDenom)
	require.NoError(t, err)

	market2, err := input.keeper.Add(input.ctx, input.baseEthDenom, input.quoteDenom)
	require.NoError(t, err)

	// get all
	{
		list := input.keeper.GetList(input.ctx)
		require.Len(t, list, 2)
		require.Equal(t, list[0].ID.UInt64(), uint64(0))
		require.Equal(t, list[1].ID.UInt64(), uint64(1))
		require.Equal(t, list[0].BaseAssetDenom, market1.BaseAssetDenom)
		require.Equal(t, list[1].BaseAssetDenom, market2.BaseAssetDenom)
		require.Equal(t, list[0].QuoteAssetDenom, market1.QuoteAssetDenom)
		require.Equal(t, list[1].QuoteAssetDenom, market2.QuoteAssetDenom)
	}

	// get filtered
	{
		// check limit
		{
			params := types.MarketsReq{
				Page:  1,
				Limit: 1,
			}
			list := input.keeper.GetListFiltered(input.ctx, params)
			require.Len(t, list, 1)
		}

		// check quote filtering
		{
			params := types.MarketsReq{
				Page:            1,
				QuoteAssetDenom: input.quoteDenom,
			}
			list := input.keeper.GetListFiltered(input.ctx, params)
			require.Len(t, list, 2)
			require.Equal(t, list[0].BaseAssetDenom, input.baseBtcDenom)
			require.Equal(t, list[1].BaseAssetDenom, input.baseEthDenom)
		}

		// check base filtering
		{
			params := types.MarketsReq{
				Page:           1,
				BaseAssetDenom: input.baseEthDenom,
			}
			list := input.keeper.GetListFiltered(input.ctx, params)
			require.Len(t, list, 1)
			require.Equal(t, list[0].BaseAssetDenom, input.baseEthDenom)
		}

		// check no-result filtering
		{
			params := types.MarketsReq{
				Page:           1,
				BaseAssetDenom: "base",
			}
			list := input.keeper.GetListFiltered(input.ctx, params)
			require.Empty(t, list)
		}
	}
}
