// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

func TestMarketsKeeper_Genesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper, cdc := input.ctx, input.keeper, input.cdc

	lastID := dnTypes.NewIDFromUint64(1)
	initState := types.GenesisState{
		Markets: types.Markets{
			{
				ID:              dnTypes.NewIDFromUint64(0),
				BaseAssetDenom:  input.baseBtcDenom,
				QuoteAssetDenom: input.quoteDenom,
			},
			{
				ID:              dnTypes.NewIDFromUint64(1),
				BaseAssetDenom:  input.baseEthDenom,
				QuoteAssetDenom: input.quoteDenom,
			},
		},
		LastMarketID: &lastID,
	}

	// init
	{
		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(initState))

		// lastID
		require.NotNil(t, keeper.getLastMarketID(ctx))
		require.Equal(t, initState.LastMarketID.String(), keeper.getLastMarketID(ctx).String())

		// markets
		getMarkets := keeper.GetList(ctx)
		require.Len(t, getMarkets, len(initState.Markets))
		for _, initMarket := range initState.Markets {
			foundCnt := 0
			for _, getMarket := range getMarkets {
				if getMarket.ID.Equal(initMarket.ID) {
					require.Equal(t, initMarket.BaseAssetDenom, getMarket.BaseAssetDenom)
					require.Equal(t, initMarket.QuoteAssetDenom, getMarket.QuoteAssetDenom)
					foundCnt++
				}
			}
			require.Equal(t, 1, foundCnt)
		}
	}

	// export
	{
		var exportState types.GenesisState
		cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportState)

		// lastID
		require.NotNil(t, exportState.LastMarketID)
		require.Equal(t, keeper.getLastMarketID(ctx).String(), exportState.LastMarketID.String())

		// markets
		getMarkets := keeper.GetList(ctx)
		require.Len(t, exportState.Markets, len(getMarkets))
		for _, getMarket := range getMarkets {
			foundCnt := 0
			for _, exportMarket := range exportState.Markets {
				if getMarket.ID.Equal(exportMarket.ID) {
					require.Equal(t, getMarket.BaseAssetDenom, exportMarket.BaseAssetDenom)
					require.Equal(t, getMarket.QuoteAssetDenom, exportMarket.QuoteAssetDenom)
					foundCnt++
				}
			}
			require.Equal(t, 1, foundCnt)
		}
	}

	// init with non-existing currency
	{
		lastID := dnTypes.NewIDFromUint64(0)
		fakeState := types.GenesisState{
			Markets: types.Markets{
				{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "test",
					QuoteAssetDenom: input.quoteDenom,
				},
			},
			LastMarketID: &lastID,
		}

		require.Panics(t, func() {
			keeper.InitGenesis(ctx, cdc.MustMarshalJSON(fakeState))
		})
	}
}
