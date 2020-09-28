// +build unit

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func TestOracleKeeper_Genesis_Init(t *testing.T) {
	input := NewTestInput(t)

	keeper := input.keeper
	ctx := input.ctx
	ctx = ctx.WithBlockTime(time.Now().Add(time.Hour))
	ctx = ctx.WithBlockHeight(3)
	cdc := input.cdc

	// default genesis
	{
		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(types.DefaultGenesisState()))
		cpList, err := keeper.GetCurrentPricesList(input.ctx)
		require.Nil(t, err)
		params := keeper.GetParams(input.ctx)
		require.Len(t, cpList, 0)
		defaultParams := types.DefaultParams()
		require.Equal(t, len(params.Assets), len(defaultParams.Assets))
	}

	// prices asset code doesnt exist in assets
	{
		func() {
			params := input.keeper.GetParams(ctx)
			params.Assets = types.Assets{
				types.NewAsset(dnTypes.AssetCode("eth_xfi"), types.Oracles{}, true),
			}

			cpList := types.CurrentPrices{
				NewMockCurrentPrice("btc_xfi", 100, 99),
			}

			state := types.GenesisState{
				Params:        params,
				CurrentPrices: cpList,
			}

			defer func() {
				r := recover()
				require.NotNil(t, r)
				require.Contains(t, r.(error).Error(), "asset_code")
				require.Contains(t, r.(error).Error(), "does not exist")
			}()

			keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))
		}()
	}

	//import and export
	{
		params := input.keeper.GetParams(ctx)

		oracles := types.Oracles{
			types.Oracle{
				Address: sdk.AccAddress{},
			},
		}

		params.Assets = types.Assets{
			types.NewAsset(dnTypes.AssetCode("btc_xfi"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("eth_xfi"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("xfi_btc"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("usdt_xfi"), oracles, true),
		}

		cpList := types.CurrentPrices{
			NewMockCurrentPrice("btc_xfi", 100, 99),
			NewMockCurrentPrice("eth_xfi", 200, 199),
			NewMockCurrentPrice("xfi_btc", 300, 298),
			NewMockCurrentPrice("usdt_xfi", 400, 389),
		}

		state := types.GenesisState{
			Params:        params,
			CurrentPrices: cpList,
		}

		// initialize and check current state with init values
		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))

		cpListFromKeeper, err := keeper.GetCurrentPricesList(ctx)
		require.Nil(t, err)
		require.Len(t, cpListFromKeeper, len(cpList))

		paramsFromKeeper := keeper.GetParams(ctx)
		require.Equal(t, paramsFromKeeper.Assets, params.Assets)

		// export and check exported values with initial
		var exportedState types.GenesisState
		cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

		require.False(t, exportedState.IsEmpty())
		require.Equal(t, exportedState.Params.Assets, params.Assets)
		require.Equal(t, exportedState.Params.Nominees, params.Nominees)
		require.Equal(t, exportedState.Params.PostPrice, params.PostPrice)
		require.Equal(t, len(exportedState.CurrentPrices), len(state.CurrentPrices))

		// checking all of items existing in the export
		sumAskPrices, sumBidPrices := sdk.NewDec(0), sdk.NewDec(0)
		for _, i := range exportedState.CurrentPrices {
			sumAskPrices = sumAskPrices.Add(i.AskPrice)
			sumBidPrices = sumBidPrices.Add(i.BidPrice)
		}

		sumAskPricesInitial, sumBidPricesInitial := sdk.NewDec(0), sdk.NewDec(0)
		for _, i := range state.CurrentPrices {
			sumAskPricesInitial = sumAskPricesInitial.Add(i.AskPrice)
			sumBidPricesInitial = sumBidPricesInitial.Add(i.BidPrice)
		}

		require.Equal(t, sumAskPrices, sumAskPricesInitial)
		require.Equal(t, sumBidPrices, sumBidPricesInitial)
	}
}
