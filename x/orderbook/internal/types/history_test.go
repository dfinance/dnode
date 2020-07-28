// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func NewMockHistoryItem(id uint64) HistoryItem {
	return HistoryItem{
		MarketID:         dnTypes.NewIDFromUint64(id),
		ClearancePrice:   sdk.NewUintFromString("100"),
		BidOrdersCount:   1,
		AskOrdersCount:   1,
		BidVolume:        sdk.NewUintFromString("200"),
		AskVolume:        sdk.NewUintFromString("200"),
		MatchedBidVolume: sdk.NewUintFromString("200"),
		MatchedAskVolume: sdk.NewUintFromString("200"),
		Timestamp:        time.Now().Unix(),
		BlockHeight:      1,
	}
}

func TestOrderBook_History_Valid(t *testing.T) {
	// ok
	{
		item := NewMockHistoryItem(1)
		require.NoError(t, item.Valid())
	}

	// fail: timestamp
	{
		item := NewMockHistoryItem(1)
		item.Timestamp = -1
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "timestamp")
		require.Contains(t, item.Valid().Error(), "negative")
	}

	// fail: marketId
	{
		item := NewMockHistoryItem(1)
		item.MarketID = dnTypes.ID(sdk.Uint{})
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "market_id")
		require.Contains(t, item.Valid().Error(), "nil")
	}

	// fail: block height
	{
		item := NewMockHistoryItem(1)
		item.BlockHeight = -1
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "block_height")
		require.Contains(t, item.Valid().Error(), "negative")
	}

	// fail: ClearancePrice is zero
	{
		item := NewMockHistoryItem(1)
		item.ClearancePrice = sdk.NewUintFromString("0")
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "clearance_price")
		require.Contains(t, item.Valid().Error(), "zero")
	}

	// fail: BidOrdersCount is zero
	{
		item := NewMockHistoryItem(1)
		item.BidOrdersCount = -1
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "bid_orders_count")
		require.Contains(t, item.Valid().Error(), "negative")
	}
	// fail: AskOrdersCount is zero
	{
		item := NewMockHistoryItem(1)
		item.AskOrdersCount = -1
		require.Error(t, item.Valid())
		require.Contains(t, item.Valid().Error(), "ask_orders_count")
		require.Contains(t, item.Valid().Error(), "negative")
	}
}
