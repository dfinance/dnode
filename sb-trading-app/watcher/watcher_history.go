package watcher

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/require"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

type History struct {
	sync.RWMutex
	t             *testing.T
	curMarketItem map[string]*HistoryItem
	MarketItems   map[string]HistoryItems
}

type HistoryItem struct {
	ClearancePrice        sdk.Uint
	PostedOrders          uint64
	CanceledOrders        uint64
	FullyFilledOrders     uint64
	PartiallyFilledOrders uint64
	AccountBalances       HistoryBalances
}

type HistoryItems []HistoryItem

type HistoryBalance struct {
	Name  string
	Base  sdk.Uint
	Quote sdk.Uint
}

type HistoryBalances []HistoryBalance

func (h *History) SetCurBalances(marketID string, balances HistoryBalances) {
	h.Lock()
	defer h.Unlock()

	if h.curMarketItem == nil {
		return
	}

	h.curMarketItem[marketID].AccountBalances = balances
}

func (h *History) ResetCurItem() {
	h.Lock()
	defer h.Unlock()

	if h.curMarketItem != nil {
		for marketID, item := range h.curMarketItem {
			h.MarketItems[marketID] = append(h.MarketItems[marketID], *item)
		}
	}

	h.curMarketItem = make(map[string]*HistoryItem, len(h.MarketItems))
	for marketID := range h.MarketItems {
		h.curMarketItem[marketID] = &HistoryItem{
			ClearancePrice: sdk.ZeroUint(),
		}
	}
}

func (h *History) String(stats, balances bool) string {
	var buf bytes.Buffer

	for marketID, items := range h.MarketItems {
		if stats {
			buf.WriteString(fmt.Sprintf("MarketID.Stats: %s\n", marketID))

			t := tablewriter.NewWriter(&buf)
			t.SetHeader(HistoryItem{}.TableHeaders())

			for i, item := range items {
				t.Append(item.TableValues(int64(i)))
			}
			t.Render()
		}

		if balances {
			buf.WriteString(fmt.Sprintf("MarketID.Balances: %s\n", marketID))

			t := tablewriter.NewWriter(&buf)
			t.SetHeader(HistoryBalance{}.TableHeaders())

			for i, item := range items {
				for _, balance := range item.AccountBalances {
					t.Append(balance.TableValues(int64(i)))
				}
			}
			t.Render()
		}
	}

	return buf.String()
}

func (h *History) HandleOrderPostEvent(event coreTypes.ResultEvent) {
	marketIDCount, ok := countEventAttrMarketIDOrders(event)
	require.True(h.t, ok, "market_id not found: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, count := range marketIDCount {
		h.curMarketItem[marketID].PostedOrders += uint64(count)
	}
}

func (h *History) HandleOrderCancelEvent(event coreTypes.ResultEvent) {
	marketIDCount, ok := countEventAttrMarketIDOrders(event)
	require.True(h.t, ok, "market_id not found: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, count := range marketIDCount {
		h.curMarketItem[marketID].CanceledOrders += uint64(count)
	}
}

func (h *History) HandleOrderFullFillEvent(event coreTypes.ResultEvent) {
	marketIDCount, ok := countEventAttrMarketIDOrders(event)
	require.True(h.t, ok, "market_id not found: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, count := range marketIDCount {
		h.curMarketItem[marketID].FullyFilledOrders += uint64(count)
	}
}

func (h *History) HandleOrderPartialFillEvent(event coreTypes.ResultEvent) {
	marketIDCount, ok := countEventAttrMarketIDOrders(event)
	require.True(h.t, ok, "market_id not found: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, count := range marketIDCount {
		h.curMarketItem[marketID].PartiallyFilledOrders += uint64(count)
	}
}

func (h *History) HandleOrderBookClearanceEvent(event coreTypes.ResultEvent) {
	marketPrices, ok := findEventAttrPrices(event)
	require.True(h.t, ok, "price not found: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, price := range marketPrices {
		h.curMarketItem[marketID].ClearancePrice = price
	}
}

func countEventAttrMarketIDOrders(event coreTypes.ResultEvent) (marketCounts map[string]int, ok bool) {
	ok = true
	marketCounts = make(map[string]int, 0)

	for eventType, eventValues := range event.Events {
		if strings.Contains(eventType, "market_id") {
			for _, marketID := range eventValues {
				marketCounts[marketID] = marketCounts[marketID] + 1
			}
			break
		}
	}
	if len(marketCounts) == 0 {
		ok = false
		return
	}

	return
}

func findEventAttrPrices(event coreTypes.ResultEvent) (marketPrices map[string]sdk.Uint, ok bool) {
	ok = true
	var marketsStr, pricesStr []string

	for eventType, eventValues := range event.Events {
		if eventType == "orderbook.clearance.market_id" {
			marketsStr = eventValues
			continue
		}
		if eventType == "orderbook.clearance.price" {
			pricesStr = eventValues
			continue
		}
	}
	if len(marketsStr) == 0 || len(pricesStr) == 0 {
		ok = false
		return
	}
	if len(marketsStr) != len(pricesStr) {
		ok = false
		return
	}

	marketPrices = make(map[string]sdk.Uint, len(pricesStr))
	for i, marketID := range marketsStr {
		price := sdk.NewUintFromString(pricesStr[i])
		marketPrices[marketID] = price
	}

	return
}

func (b HistoryBalance) TableHeaders() []string {
	h := []string{
		"ID",
		"Name",
		"Base",
		"Quote",
	}

	return h
}

func (b HistoryBalance) TableValues(id int64) []string {
	v := []string{
		strconv.FormatInt(id, 10),
		b.Name,
		b.Base.String(),
		b.Quote.String(),
	}

	return v
}

func (i HistoryItem) TableHeaders() []string {
	h := []string{
		"ID",
		"Price",
		"Posts",
		"Cancels",
		"FFills",
		"PFills",
	}

	return h
}

func (i HistoryItem) TableValues(id int64) []string {
	v := []string{
		strconv.FormatInt(id, 10),
		i.ClearancePrice.String(),
		strconv.FormatUint(i.PostedOrders, 10),
		strconv.FormatUint(i.CanceledOrders, 10),
		strconv.FormatUint(i.FullyFilledOrders, 10),
		strconv.FormatUint(i.PartiallyFilledOrders, 10),
	}

	return v
}

func NewHistory(t *testing.T, marketIDs []string) *History {
	h := &History{
		t:           t,
		MarketItems: make(map[string]HistoryItems, len(marketIDs)),
	}

	for _, id := range marketIDs {
		h.MarketItems[id] = make(HistoryItems, 0)
	}

	h.ResetCurItem()

	return h
}
