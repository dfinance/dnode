package watcher

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/require"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/dfinance/dnode/x/currencies_register"
)

type History struct {
	sync.RWMutex
	t                   *testing.T
	curMarketItem       map[string]*HistoryItem
	prevBlockTS         time.Time
	curBots             uint
	MarketInfos         map[string]MarketInfo      // key: marketID
	MarketItems         map[string]HistoryItems    // key: marketID
	CountedPostOrders   map[string]bool            // key: orderID
	CountedCancelOrders map[string]bool            // key: orderID
	CountedFFillOrders  map[string]bool            // key: orderID
	CountedPFillOrders  map[string]map[string]bool // key1: orderID, key2: quantity
	BlockInfos          []BlockInfo
}

type BlockInfo struct {
	Clients  uint
	Duration time.Duration
}

type MarketInfo struct {
	BaseCurrency  currencies_register.CurrencyInfo
	QuoteCurrency currencies_register.CurrencyInfo
}

type HistoryItem struct {
	Clients               uint
	ClearancePrice        sdk.Uint
	PostedOrders          uint64
	CanceledOrders        uint64
	FullyFilledOrders     uint64
	PartiallyFilledOrders uint64
	AccountBalances       HistoryBalances
	Timestamp             time.Time
	PriceToDec            func(price sdk.Uint) sdk.Dec
}

type HistoryItems []HistoryItem

type HistoryBalance struct {
	Name  string
	Base  sdk.Uint
	Quote sdk.Uint
}

type HistoryBalances []HistoryBalance

func (h *History) SetCurBots(count uint) {
	h.Lock()
	defer h.Unlock()

	h.curBots = count
}

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
			item.Clients = h.curBots
			h.MarketItems[marketID] = append(h.MarketItems[marketID], *item)
		}
	}

	now := time.Now()
	h.curMarketItem = make(map[string]*HistoryItem, len(h.MarketItems))
	for marketID := range h.MarketItems {
		h.curMarketItem[marketID] = &HistoryItem{
			Timestamp:      now,
			ClearancePrice: sdk.ZeroUint(),
			PriceToDec:     h.MarketInfos[marketID].QuoteCurrency.UintToDec,
		}
	}
}

func (h *History) String(stats, balances, blockDurations bool) string {
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

	if blockDurations {
		buf.WriteString("Block durations:\n")

		t := tablewriter.NewWriter(&buf)
		t.SetHeader([]string{
			"#",
			"Clients",
			"Duration",
		})
		for i, info := range h.BlockInfos {
			t.Append([]string{
				strconv.FormatInt(int64(i), 10),
				strconv.FormatUint(uint64(info.Clients), 10),
				info.Duration.String(),
			})
		}
		t.Render()
	}

	return buf.String()
}

func (h *History) HandleOrderPostEvent(event coreTypes.ResultEvent) {
	marketOrders := findEventAttrMarketIDOrders("orders.post.", event)
	require.NotZero(h.t, len(marketOrders), "parsing failed: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, orderIDs := range marketOrders {
		for _, orderID := range orderIDs {
			if _, ok := h.CountedPostOrders[orderID]; !ok {
				h.curMarketItem[marketID].PostedOrders++
				h.CountedPostOrders[orderID] = true
			}
		}
	}
}

func (h *History) HandleOrderCancelEvent(event coreTypes.ResultEvent) {
	marketOrders := findEventAttrMarketIDOrders("orders.cancel.", event)
	require.NotZero(h.t, len(marketOrders), "parsing failed: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, orderIDs := range marketOrders {
		for _, orderID := range orderIDs {
			if _, ok := h.CountedCancelOrders[orderID]; !ok {
				h.curMarketItem[marketID].CanceledOrders++
				h.CountedCancelOrders[orderID] = true
			}

			if _, ok := h.CountedPostOrders[orderID]; !ok {
				h.curMarketItem[marketID].PostedOrders++
				h.CountedPostOrders[orderID] = true
			}
		}
	}
}

func (h *History) HandleOrderFullFillEvent(event coreTypes.ResultEvent) {
	marketOrders := findEventAttrMarketIDOrders("orders.full_fill.", event)
	require.NotZero(h.t, len(marketOrders), "parsing failed: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, orderIDs := range marketOrders {
		for _, orderID := range orderIDs {
			if _, ok := h.CountedFFillOrders[orderID]; !ok {
				h.curMarketItem[marketID].FullyFilledOrders++
				h.CountedFFillOrders[orderID] = true
			}

			if _, ok := h.CountedPostOrders[orderID]; !ok {
				h.curMarketItem[marketID].PostedOrders++
				h.CountedPostOrders[orderID] = true
			}
		}
	}
}

func (h *History) HandleOrderPartialFillEvent(event coreTypes.ResultEvent) {
	marketOrders, ordersQuantity := findEventAttrMarketIDOrdersQuantity("orders.partial_fill.", event)
	require.NotZero(h.t, len(marketOrders), "parsing failed: %v", event)

	h.Lock()
	defer h.Unlock()

	for marketID, orderIDs := range marketOrders {
		for _, orderID := range orderIDs {
			if _, ok := h.CountedPFillOrders[orderID]; !ok {
				h.CountedPFillOrders[orderID] = make(map[string]bool, 0)
			}

			quantity := ordersQuantity[orderID]
			if _, ok := h.CountedPFillOrders[orderID][quantity.String()]; !ok {
				h.curMarketItem[marketID].PartiallyFilledOrders++
				h.CountedPFillOrders[orderID][quantity.String()] = true
			}

			if _, ok := h.CountedPostOrders[orderID]; !ok {
				h.curMarketItem[marketID].PostedOrders++
				h.CountedPostOrders[orderID] = true
			}
		}
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

func (h *History) HandleNewBlockEvent(event coreTypes.ResultEvent) {
	curBlockTS := time.Now()

	h.Lock()
	defer h.Unlock()

	if !h.prevBlockTS.IsZero() {
		h.BlockInfos = append(h.BlockInfos, BlockInfo{
			Clients:  h.curBots,
			Duration: curBlockTS.Sub(h.prevBlockTS),
		})
	}

	h.prevBlockTS = curBlockTS
}

func findEventAttrMarketIDOrders(prefix string, event coreTypes.ResultEvent) (marketOrders map[string][]string) {
	marketOrders = make(map[string][]string, 0)

	eventTypeMarketQuery := prefix + "market_id"
	eventTypeOrderQuery := prefix + "order_id"
	marketValues := event.Events[eventTypeMarketQuery]
	orderValues := event.Events[eventTypeOrderQuery]

	if len(marketValues) == 0 || len(orderValues) == 0 {
		return
	}
	if len(marketValues) != len(orderValues) {
		return
	}

	for i := 0; i < len(marketValues); i++ {
		marketID := marketValues[i]
		orderID := orderValues[i]
		marketOrders[marketID] = append(marketOrders[marketID], orderID)
	}

	return
}

func findEventAttrMarketIDOrdersQuantity(prefix string, event coreTypes.ResultEvent) (marketOrders map[string][]string, ordersQuantity map[string]sdk.Uint) {
	marketOrders = make(map[string][]string, 0)
	ordersQuantity = make(map[string]sdk.Uint, 0)

	eventTypeMarketQuery := prefix + "market_id"
	eventTypeOrderQuery := prefix + "order_id"
	eventTypeQuantityQuery := prefix + "quantity"
	marketValues := event.Events[eventTypeMarketQuery]
	orderValues := event.Events[eventTypeOrderQuery]
	quantityValues := event.Events[eventTypeQuantityQuery]

	if len(marketValues) == 0 || len(orderValues) == 0 || len(quantityValues) == 0 {
		return
	}
	if len(marketValues) != len(orderValues) && len(marketValues) != len(quantityValues) {
		return
	}

	for i := 0; i < len(marketValues); i++ {
		marketID := marketValues[i]
		orderID := orderValues[i]
		quantity := sdk.NewUintFromString(quantityValues[i])

		marketOrders[marketID] = append(marketOrders[marketID], orderID)
		ordersQuantity[orderID] = quantity
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
		"StartedAt",
		"Clients",
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
		i.Timestamp.UTC().String(),
		strconv.FormatUint(uint64(i.Clients), 10),
		i.PriceToDec(i.ClearancePrice).String(),
		strconv.FormatUint(i.PostedOrders, 10),
		strconv.FormatUint(i.CanceledOrders, 10),
		strconv.FormatUint(i.FullyFilledOrders, 10),
		strconv.FormatUint(i.PartiallyFilledOrders, 10),
	}

	return v
}

func NewHistory(t *testing.T, markets map[string]MarketInfo, startBots uint) *History {
	h := &History{
		t:                   t,
		curBots:             startBots,
		MarketInfos:         make(map[string]MarketInfo, len(markets)),
		MarketItems:         make(map[string]HistoryItems, len(markets)),
		CountedPostOrders:   make(map[string]bool, 0),
		CountedCancelOrders: make(map[string]bool, 0),
		CountedFFillOrders:  make(map[string]bool, 0),
		CountedPFillOrders:  make(map[string]map[string]bool, 0),
	}

	for id, info := range markets {
		h.MarketInfos[id] = info
		h.MarketItems[id] = make(HistoryItems, 0)
	}

	h.ResetCurItem()

	return h
}
