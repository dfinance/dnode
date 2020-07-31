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
		params := input.keeper.GetParams(ctx)
		params.Assets = types.Assets{
			types.NewAsset(dnTypes.AssetCode("eth_dfi"), types.Oracles{}, true),
		}

		cpList := types.CurrentPrices{
			NewMockCurrentPrice("btc_dfi", 100),
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
			types.NewAsset(dnTypes.AssetCode("btc_dfi"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("eth_dfi"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("dfi_btc"), oracles, true),
			types.NewAsset(dnTypes.AssetCode("usdt_dfi"), oracles, true),
		}

		cpList := types.CurrentPrices{
			NewMockCurrentPrice("btc_dfi", 100),
			NewMockCurrentPrice("eth_dfi", 200),
			NewMockCurrentPrice("dfi_btc", 300),
			NewMockCurrentPrice("usdt_dfi", 400),
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
		sumPrices := sdk.NewIntFromUint64(0)
		for _, i := range exportedState.CurrentPrices {
			sumPrices = sumPrices.Add(i.Price)
		}

		sumPricesInitial := sdk.NewIntFromUint64(0)
		for _, i := range state.CurrentPrices {
			sumPricesInitial = sumPricesInitial.Add(i.Price)
		}

		require.Equal(t, sumPrices, sumPricesInitial)
	}
}
