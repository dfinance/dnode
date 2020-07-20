// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

func CompareHistoryItems(t *testing.T, item1, item2 types.HistoryItem) {
	require.True(t, item1.MarketID.Equal(item2.MarketID), "MarketID")
	require.True(t, item1.ClearancePrice.Equal(item2.ClearancePrice), "ClearancePrice")
	require.Equal(t, item1.BidOrdersCount, item2.BidOrdersCount, "BidOrdersCount")
	require.Equal(t, item1.AskOrdersCount, item2.AskOrdersCount, "AskOrdersCount")
	require.True(t, item1.BidVolume.Equal(item2.BidVolume), "BidVolume")
	require.True(t, item1.AskVolume.Equal(item2.AskVolume), "AskVolume")
	require.True(t, item1.MatchedBidVolume.Equal(item2.MatchedBidVolume), "MatchedBidVolume")
	require.True(t, item1.MatchedAskVolume.Equal(item2.MatchedAskVolume), "MatchedAskVolume")
	require.Equal(t, item1.Timestamp, item2.Timestamp, "Timestamp")
	require.Equal(t, item1.BlockHeight, item2.BlockHeight, "BlockHeight")
}

func TestOBKeeper_History_StoreIO(t *testing.T) {
	input := NewTestInput(t)
	marketID := dnTypes.NewIDFromUint64(0)

	// non-existing item
	{
		blockHeight := int64(1)
		require.False(t, input.keeper.HasHistoryItem(input.ctx, marketID, blockHeight))

		_, err := input.keeper.GetHistoryItem(input.ctx, marketID, blockHeight)
		require.Error(t, err)
	}

	inputItem1 := NewMockHistoryItem(marketID, 1)
	inputItem2 := NewMockHistoryItem(marketID, 2)
	inputItem3 := NewMockHistoryItem(marketID, 3)
	inputItems := types.HistoryItems{inputItem1, inputItem2, inputItem3}

	// add history items
	for i, inputItem := range inputItems {
		blockHeight := int64(i)+1
		input.keeper.SetHistoryItem(input.ctx, inputItem)
		require.True(t, input.keeper.HasHistoryItem(input.ctx, marketID, blockHeight))

		outputItem, err := input.keeper.GetHistoryItem(input.ctx, marketID, blockHeight)
		require.NoError(t, err)
		CompareHistoryItems(t, inputItem, outputItem)
	}

	// check list with non-existing marketID
	{
		marketID := dnTypes.NewIDFromUint64(1)
		outputItems, err := input.keeper.GetHistoryItemsInBlockHeightRange(input.ctx, marketID, 1, 3)
		require.NoError(t, err)
		require.Len(t, outputItems, 0)
	}

	// check list with existing marketID (wrong blockHeight range)
	{
		outputItems, err := input.keeper.GetHistoryItemsInBlockHeightRange(input.ctx, marketID, 5, 10)
		require.NoError(t, err)
		require.Len(t, outputItems, 0)
	}

	// check list with existing marketID (correct blockHeight range)
	{
		outputItems, err := input.keeper.GetHistoryItemsInBlockHeightRange(input.ctx, marketID, 1, 3)
		require.NoError(t, err)
		require.Len(t, outputItems, 3)

		for i, inputItem := range inputItems {
			CompareHistoryItems(t, inputItem, outputItems[i])
		}
	}

	// check list with existing marketID (mixed blockHeight range)
	{
		outputItems, err := input.keeper.GetHistoryItemsInBlockHeightRange(input.ctx, marketID, 2, 5)
		require.NoError(t, err)
		require.Len(t, outputItems, 2)

		for i, inputItem := range inputItems[1:] {
			CompareHistoryItems(t, inputItem, outputItems[i])
		}
	}
}
