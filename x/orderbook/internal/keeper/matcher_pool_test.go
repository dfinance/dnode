// +build unit

package keeper

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/logger"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	crTypes "github.com/dfinance/dnode/x/currencies_register"
	marketTypes "github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

type MatchingPoolInput struct {
	Markets         []MatchingPoolMarketInput
	Orders          []MatchingPoolOrderInput
	bidMarketOrders map[uint64]int
	askMarketOrders map[uint64]int
}

type MatchingPoolMarketInput struct {
	BaseDenom     string
	QuoteDenom    string
	BaseDecimals  uint8
	QuoteDecimals uint8
}

type MatchingPoolOrderInput struct {
	MarketID  uint64
	OrderID   uint64
	Direction orderTypes.Direction
	Price     uint64
	// Order initial quantity.
	InQuantity uint64
	// Order quantity after match (0 - don't check).
	OutQuantity uint64
	// Order fill sequence number (0 - don't check, global between markets).
	OutFillSeq uint64
}

func (i *MatchingPoolInput) PostOrders(t *testing.T, pool *MatcherPool) {
	placedOrders := make(map[uint64]bool)
	placedFillOrders := make(map[uint64]bool)
	i.bidMarketOrders = make(map[uint64]int)
	i.askMarketOrders = make(map[uint64]int)

	extMarkets := make([]marketTypes.MarketExtended, 0, len(i.Markets))
	for id, input := range i.Markets {
		baseCurrency := crTypes.CurrencyInfo{Denom: []byte(input.BaseDenom), Decimals: input.BaseDecimals, IsToken: false, Owner: nil, TotalSupply: nil}
		quoteCurrency := crTypes.CurrencyInfo{Denom: []byte(input.QuoteDenom), Decimals: input.QuoteDecimals, IsToken: false, Owner: nil, TotalSupply: nil}
		market := marketTypes.NewMarket(dnTypes.NewIDFromUint64(uint64(id)), input.BaseDenom, input.QuoteDenom)
		marketExt := marketTypes.NewMarketExtended(market, baseCurrency, quoteCurrency)
		extMarkets = append(extMarkets, marketExt)
	}

	for _, input := range i.Orders {
		order := orderTypes.Order{
			ID:        dnTypes.NewIDFromUint64(input.OrderID),
			Market:    extMarkets[input.MarketID],
			Direction: input.Direction,
			Price:     sdk.NewUint(input.Price),
			Quantity:  sdk.NewUint(input.InQuantity),
		}

		if err := pool.AddOrder(order); err != nil {
			t.Fatalf("AddOrder: %v", err)
		}

		if order.Direction == orderTypes.BidDirection {
			i.bidMarketOrders[input.MarketID]++
		} else {
			i.askMarketOrders[input.MarketID]++
		}

		require.False(t, placedOrders[input.OrderID], "duplicate orderID %d found", input.OrderID)
		placedOrders[input.OrderID] = true

		if input.OutFillSeq > 0 {
			require.False(t, placedFillOrders[input.OutFillSeq], "duplicate order fill sequence number %d found", input.OutFillSeq)
			placedFillOrders[input.OutFillSeq] = true
		}
	}
}

func (i *MatchingPoolInput) Check(t *testing.T, results types.MatcherResults) {
	require.Len(t, results, len(i.Markets))

	for _, result := range results {
		marketID := result.MarketID.UInt64()
		require.Equal(t, result.BidOrdersCount, i.bidMarketOrders[marketID], "market %d: bidOrders count", marketID)
		require.Equal(t, result.AskOrdersCount, i.askMarketOrders[marketID], "market %d: askOrders count", marketID)

		for fillSeqNumber, fill := range result.OrderFills {
			orderInputFound := false
			for _, orderInput := range i.Orders {
				if orderInput.OrderID == fill.Order.ID.UInt64() {
					orderInputFound = true
				} else {
					continue
				}
				if orderInput.OutQuantity == 0 {
					continue
				}

				require.Equal(t, orderInput.OutQuantity, fill.QuantityFilled.Uint64(), "market %d, order %d: OutQuantity / FillQuantity", orderInput.MarketID, orderInput.OrderID)
				require.Equal(t, orderInput.InQuantity-orderInput.OutQuantity, fill.QuantityUnfilled.Uint64(), "market %d, order %d: (OutQ - InQ) / UnFillQuantity", orderInput.MarketID, orderInput.OrderID)
				if orderInput.OutFillSeq > 0 {
					require.Equal(t, orderInput.OutFillSeq, fillSeqNumber+1, "market %d, order %d: invalid fill sequence number", orderInput.MarketID, orderInput.OrderID)
				}
				break
			}
			require.True(t, orderInputFound, "fill for order %d: input not found", fill.Order.ID.UInt64())
		}
	}
}

func (i *MatchingPoolInput) PrintResults(results types.MatcherResults) {
	for _, result := range results {
		fmt.Printf("Market %d: %s\n", result.MarketID.UInt64(), result.ClearanceState.String())
	}

	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader([]string{
		"Market",
		"Order",
		"Dir",
		"Price",
		"Quantity",
		"Fill",
		"UnFill",
		"FillSeq#",
	})

	input := *i
	sort.Slice(input.Orders, func(i, j int) bool {
		iItem, jItem := input.Orders[i], input.Orders[j]

		if iItem.MarketID != jItem.MarketID {
			if iItem.MarketID > jItem.MarketID {
				return false
			}
			return true
		}

		if iItem.Direction != jItem.Direction {
			if iItem.MarketID == jItem.MarketID {
				if iItem.Direction == orderTypes.BidDirection {
					return false
				}
				return true
			}
		}

		if iItem.MarketID == jItem.MarketID {
			if iItem.Price == jItem.Price {
				if iItem.OrderID > jItem.OrderID {
					return false
				}
			} else {
				if iItem.Price < jItem.Price {
					return false
				}
			}
			return true
		}

		return true
	})

	for _, orderInput := range input.Orders {
		fillSeqNumber, fillQuantity, unFillQuantity := int64(-1), uint64(0), uint64(0)
		for _, result := range results {
			found := false
			for fillIdx, fill := range result.OrderFills {
				if fill.Order.ID.UInt64() == orderInput.OrderID {
					found = true
					fillSeqNumber = int64(fillIdx) + 1
					fillQuantity = fill.QuantityFilled.Uint64()
					unFillQuantity = fill.QuantityUnfilled.Uint64()
					break
				}
			}
			if found {
				break
			}
		}

		tableValues := []string{
			strconv.FormatUint(orderInput.MarketID, 10),
			strconv.FormatUint(orderInput.OrderID, 10),
			orderInput.Direction.String(),
			strconv.FormatUint(orderInput.Price, 10),
			strconv.FormatUint(orderInput.InQuantity, 10),
			strconv.FormatUint(fillQuantity, 10),
			strconv.FormatUint(unFillQuantity, 10),
		}
		if fillSeqNumber == -1 {
			tableValues = append(tableValues, "")
		} else {
			tableValues = append(tableValues, strconv.FormatInt(fillSeqNumber, 10))
		}
		t.Append(tableValues)
	}
	t.Render()

	fmt.Println(buf.String())
}

func (i *MatchingPoolInput) PrintCurves(pool *MatcherPool) {
	for marketID := range i.Markets {
		sdCurves := pool.GetSDCurves(dnTypes.NewIDFromUint64(uint64(marketID)))
		fmt.Printf("Market %d: SDCurves:\n", marketID)
		if sdCurves == nil {
			fmt.Println("nil")
		} else {
			fmt.Println(sdCurves.Graph())
		}
	}
}

func Test_Matching_XARExample(t *testing.T) {
	// End-to-end test with XAR example.
	inputs := MatchingPoolInput{
		Markets: []MatchingPoolMarketInput{
			{BaseDenom: "btc", QuoteDenom: "dfi", BaseDecimals: 0, QuoteDecimals: 0},
		},
		Orders: []MatchingPoolOrderInput{
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 5, Price: 12, InQuantity: 100, OutQuantity: 100},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 6, Price: 10, InQuantity: 50, OutQuantity: 42},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 4, Price: 10, InQuantity: 50, OutQuantity: 42},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 3, Price: 12, InQuantity: 100, OutQuantity: 100},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 2, Price: 8, InQuantity: 100, OutQuantity: 84},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 10, Price: 14, InQuantity: 100, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 7, Price: 14, InQuantity: 50, OutQuantity: 50},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 8, Price: 11, InQuantity: 100, OutQuantity: 82},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 1, Price: 10, InQuantity: 100, OutQuantity: 0},
		},
	}

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	inputs.PostOrders(t, &matcherPool)
	results := matcherPool.Process()
	inputs.Check(t, results)
	inputs.PrintResults(results)
	inputs.PrintCurves(&matcherPool)
}

func Test_Matching_NotionExample(t *testing.T) {
	// End-to-end test with Dfinance Notion example.
	inputs := MatchingPoolInput{
		Markets: []MatchingPoolMarketInput{
			{BaseDenom: "btc", QuoteDenom: "dfi", BaseDecimals: 0, QuoteDecimals: 0},
		},
		Orders: []MatchingPoolOrderInput{
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 1, Price: 50, InQuantity: 150, OutQuantity: 101},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 2, Price: 40, InQuantity: 90, OutQuantity: 61},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 3, Price: 55, InQuantity: 80, OutQuantity: 80},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 4, Price: 70, InQuantity: 50, OutQuantity: 50},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 5, Price: 30, InQuantity: 150, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 6, Price: 70, InQuantity: 200, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 7, Price: 55, InQuantity: 100, OutQuantity: 100},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 8, Price: 40, InQuantity: 100, OutQuantity: 68},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 9, Price: 20, InQuantity: 200, OutQuantity: 0},
		},
	}

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	inputs.PostOrders(t, &matcherPool)
	results := matcherPool.Process()
	inputs.Check(t, results)
	inputs.PrintResults(results)
	inputs.PrintCurves(&matcherPool)
}

func Test_Matching_ProRata(t *testing.T) {
	// Specific test.
	// Ask amount is twice larger than Bid amount (ProRata = 2.0): only half of Asks should be filled.
	inputs := MatchingPoolInput{
		Markets: []MatchingPoolMarketInput{
			{BaseDenom: "btc", QuoteDenom: "dfi", BaseDecimals: 0, QuoteDecimals: 0},
		},
		Orders: []MatchingPoolOrderInput{
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 7, Price: 50, InQuantity: 100, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 5, Price: 50, InQuantity: 50, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 6, Price: 50, InQuantity: 50, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 0, Price: 50, InQuantity: 50, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 1, Price: 50, InQuantity: 25, OutQuantity: 0},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 2, Price: 50, InQuantity: 25, OutQuantity: 0},
		},
	}

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	inputs.PostOrders(t, &matcherPool)
	results := matcherPool.Process()
	inputs.Check(t, results)
	inputs.PrintResults(results)
}

func Test_Matching_FillPriority(t *testing.T) {
	// Specific test.
	// Result MaxAskVolume is 25 (ProRata = 3.6).
	// Ask orders would be filled from the lowest (50) and each fill would be weighted with ProRata.
	// Test checks orders fill priority (order with higher ID and equal Price is filled later).
	// Order #4 would be filled the last and MatchedAskVolume by then would be around 24.
	inputs := MatchingPoolInput{
		Markets: []MatchingPoolMarketInput{
			{BaseDenom: "btc", QuoteDenom: "dfi", BaseDecimals: 0, QuoteDecimals: 0},
		},
		Orders: []MatchingPoolOrderInput{
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 0, Price: 50, InQuantity: 25, OutFillSeq: 2},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 1, Price: 50, InQuantity: 5, OutFillSeq: 3},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 2, Price: 75, InQuantity: 30, OutFillSeq: 4},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 3, Price: 75, InQuantity: 20, OutFillSeq: 5},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 4, Price: 75, InQuantity: 10, OutFillSeq: 6},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 5, Price: 100, InQuantity: 50},
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 6, Price: 150, InQuantity: 50},

			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 10, Price: 40, InQuantity: 75},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 11, Price: 65, InQuantity: 55},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 12, Price: 120, InQuantity: 25, OutFillSeq: 1},
		},
	}

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	inputs.PostOrders(t, &matcherPool)
	results := matcherPool.Process()
	inputs.Check(t, results)
	inputs.PrintResults(results)
	inputs.PrintCurves(&matcherPool)
}

func Test_Matching_TwoOrders(t *testing.T) {
	// Two orders, no crossing check.
	// Aggregates will "draw" two parallel lines (with the same Quantity) and CP would be found by min diff (left).
	inputs := MatchingPoolInput{
		Markets: []MatchingPoolMarketInput{
			{BaseDenom: "btc", QuoteDenom: "dfi", BaseDecimals: 0, QuoteDecimals: 0},
		},
		Orders: []MatchingPoolOrderInput{
			{MarketID: 0, Direction: orderTypes.AskDirection, OrderID: 0, Price: 50, InQuantity: 50, OutQuantity: 50},
			{MarketID: 0, Direction: orderTypes.BidDirection, OrderID: 1, Price: 100, InQuantity: 50, OutQuantity: 50},
		},
	}

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	inputs.PostOrders(t, &matcherPool)
	results := matcherPool.Process()
	inputs.Check(t, results)
	inputs.PrintResults(results)
	inputs.PrintCurves(&matcherPool)
}
