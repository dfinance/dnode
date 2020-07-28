// +build unit

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

func TestOrderBookKeeper_Genesis_Init(t *testing.T) {
	input := NewTestInput(t)

	keeper := input.keeper
	ctx := input.ctx
	ctx = ctx.WithBlockTime(time.Now().Add(time.Hour))
	ctx = ctx.WithBlockHeight(3)
	cdc := input.cdc

	// default genesis
	{
		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(types.DefaultGenesisState()))
		orders, err := keeper.GetHistoryItemsList(input.ctx)
		require.Nil(t, err)
		require.Len(t, orders, 0)
	}

	// import and export
	{
		items := types.HistoryItems{
			NewMockHistoryItem(dnTypes.NewIDFromUint64(1), 1),
			NewMockHistoryItem(dnTypes.NewIDFromUint64(2), 1),
			NewMockHistoryItem(dnTypes.NewIDFromUint64(1), 2),
			NewMockHistoryItem(dnTypes.NewIDFromUint64(3), 3),
		}

		state := types.GenesisState{
			HistoryItems: items,
		}

		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))
		itemsFromKeeper, err := keeper.GetHistoryItemsList(ctx)
		require.Nil(t, err)
		require.Len(t, itemsFromKeeper, len(items))

		var exportedState types.GenesisState
		cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

		require.False(t, exportedState.IsEmpty())
		require.Equal(t, len(exportedState.HistoryItems), len(state.HistoryItems))

		sumClearancePrice := sdk.NewUint(0)
		for _, i := range exportedState.HistoryItems {
			sumClearancePrice = sumClearancePrice.Add(i.ClearancePrice)
		}

		sumClearancePriceInitial := sdk.NewUint(0)
		for _, i := range state.HistoryItems {
			sumClearancePriceInitial = sumClearancePriceInitial.Add(i.ClearancePrice)
		}
		require.Equal(t, sumClearancePriceInitial, sumClearancePrice)

	}
}
