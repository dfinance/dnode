package keeper

import (
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

// Matcher object defined for every market.
// Object builds DSCurves and fills orders using ClearanceState.
type Matcher struct {
	marketID   dnTypes.ID
	logger     log.Logger
	orders     MatcherOrders
	aggregates MatcherAggregates
	sdCurves   SDCurves
}

// MatcherOrders stores bid/ask orders in sorted slices.
type MatcherOrders struct {
	bid orderTypes.Orders
	ask orderTypes.Orders
}

// MatcherAggregates stores bid/ask aggregates.
type MatcherAggregates struct {
	bid OrderAggregates
	ask OrderAggregates
}

// AddOrder validates the input order and adds it to the corresponding queue.
func (m *Matcher) AddOrder(order *orderTypes.Order) error {
	const MaxUint = ^uint(0)
	const MaxInt = int(MaxUint >> 1)

	if order == nil {
		return sdkErrors.Wrap(types.ErrInternal, "nil order")
	}
	if !order.Market.ID.Equal(m.marketID) {
		return sdkErrors.Wrap(types.ErrInternal, "marketID mismatch")
	}
	if order.Price.IsZero() {
		return sdkErrors.Wrap(types.ErrInternal, "zero price")
	}
	if order.Quantity.IsZero() {
		return sdkErrors.Wrap(types.ErrInternal, "zero quantity")
	}

	var orders *orderTypes.Orders
	switch order.Direction {
	case orderTypes.BidDirection:
		orders = &m.orders.bid
	case orderTypes.AskDirection:
		orders = &m.orders.ask
	default:
		return fmt.Errorf("unknown order direction: %s", order.Direction)
	}

	// =) Check added just for fun
	if len(*orders) == MaxInt {
		return fmt.Errorf("max orders len reached")
	}

	*orders = append(*orders, *order)

	return nil
}

// Match sorts order queues, builds order aggregates and SDCurves.
func (m *Matcher) Match() (result types.MatcherResult, retErr error) {
	// orders sorting and aggregating (that can be safely paralleled)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		// orders should be sorted by price for aggregates build
		// orders would be filled up in reverse order, so orders came first (lower ID) should be executed firstly
		sort.Sort(ByPriceAscIDDesc(m.orders.bid))
		m.aggregates.bid = NewBidOrderAggregates(m.orders.bid)
	}()

	go func() {
		defer wg.Done()
		// orders should be sorted by price for aggregates build
		// orders would be filled up in direct order, so orders came first (lower ID) should be executed firstly
		sort.Sort(ByPriceAscIDAsc(m.orders.ask))
		m.aggregates.ask = NewAskOrderAggregates(m.orders.ask)
	}()

	wg.Wait()

	// build supply-demand curves using orders aggregates
	sdCurves, err := NewSDCurves(m.aggregates.ask, m.aggregates.bid)
	if err != nil {
		retErr = err
		return
	}
	m.sdCurves = sdCurves

	// get clearance price, max volumes for the next steps
	clearanceState, err := sdCurves.GetClearanceState()
	if err != nil {
		// this should not happen
		// if SDCurve was build with no error, crossing point (any quality) must be found
		m.logger.Debug(fmt.Sprintf("Matcher intermediate results for marketID %s:", m.marketID.String()))
		m.logger.Debug("Bid orders:")
		m.logger.Debug("\n" + m.orders.bid.String())
		m.logger.Debug("Ask orders:")
		m.logger.Debug("\n" + m.orders.ask.String())
		m.logger.Debug("PQ curves:")
		m.logger.Debug("\n" + sdCurves.String())
		retErr = err
		return
	}

	// fill up orders
	bidFills, bidMatchedVolume := m.getBidOrderFills(clearanceState)
	askFills, askMatchedVolume := m.getAskOrderFills(clearanceState)

	// build the result
	result = types.MatcherResult{
		MarketID:         m.marketID,
		ClearanceState:   clearanceState,
		BidOrdersCount:   len(m.orders.bid),
		AskOrdersCount:   len(m.orders.ask),
		MatchedBidVolume: bidMatchedVolume,
		MatchedAskVolume: askMatchedVolume,
		OrderFills:       append(bidFills, askFills...),
	}

	// TODO: should be removed later as even having debug log level off, building strings takes time
	// m.logger.Debug(fmt.Sprintf("Matcher results for marketID %s:", m.marketID.String()))
	// m.logger.Debug("Bid orders:")
	// m.logger.Debug("\n" + m.orders.bid.String())
	// m.logger.Debug("Bid aggregates:")
	// m.logger.Debug("\n" + m.aggregates.bid.String())
	// m.logger.Debug("Ask orders:")
	// m.logger.Debug("\n" + m.orders.ask.String())
	// m.logger.Debug("Ask aggregates:")
	// m.logger.Debug("\n" + m.aggregates.ask.String())
	// m.logger.Debug("PQ curves:")
	// m.logger.Debug("\n" + sdCurves.String())
	// m.logger.Debug("\n" + result.String())

	return
}

// getBidOrderFills fills up bid orders in reverse order (from highest target price and lower order IDs).
func (m *Matcher) getBidOrderFills(clearanceState types.ClearanceState) (fills orderTypes.OrderFills, matchedVolume sdk.Dec) {
	// fills stores result order fills
	// matchedVolume stores current matched volume (should be <= clearanceState.MaxBidVolume
	fills, matchedVolume = make(orderTypes.OrderFills, 0, len(m.orders.bid)), sdk.ZeroDec()

	proRataGTOne := clearanceState.ProRata.GT(sdk.OneDec())
	for i := len(m.orders.bid) - 1; i >= 0; i-- {
		order := &m.orders.bid[i]

		// stop the processing if matched volume has reached its max or orders can't filled
		if order.Price.LT(clearanceState.Price) || matchedVolume.Equal(clearanceState.MaxBidVolume) {
			break
		}

		// adjust the fill quantity using ProRata concept (proportionate bids/asks execution if demand/supply are not equal)
		fillQuantity := order.Quantity
		if !proRataGTOne {
			orderQtyDec := sdk.NewDecFromBigInt(fillQuantity.BigInt())
			orderQtyDec = orderQtyDec.Mul(clearanceState.ProRata)
			orderQtyDec = orderQtyDec.Ceil()

			if matchedVolume.Add(orderQtyDec).GT(clearanceState.MaxBidVolume) {
				orderQtyDec = clearanceState.MaxBidVolume.Sub(matchedVolume)
			}
			fillQuantity = sdk.NewUintFromBigInt(orderQtyDec.RoundInt().BigInt())
		}

		if fillQuantity.IsZero() {
			continue
		}
		matchedVolume = matchedVolume.Add(sdk.NewDecFromBigInt(fillQuantity.BigInt()))

		fills = append(fills, orderTypes.OrderFill{
			Order:            *order,
			ClearancePrice:   clearanceState.Price,
			QuantityFilled:   fillQuantity,
			QuantityUnfilled: order.Quantity.Sub(fillQuantity),
		})
	}

	return
}

// getAskOrderFills fills up ask orders in direct order (from lowest target price and lower order IDs).
func (m *Matcher) getAskOrderFills(clearanceState types.ClearanceState) (fills orderTypes.OrderFills, matchedVolume sdk.Dec) {
	// fills stores result order fills
	// matchedVolume stores current matched volume (should be <= clearanceState.MaxBidVolume
	fills, matchedVolume = make(orderTypes.OrderFills, 0, len(m.orders.ask)), sdk.ZeroDec()

	proRataGTOne := clearanceState.ProRata.GT(sdk.OneDec())
	for i := 0; i < len(m.orders.ask); i++ {
		order := &m.orders.ask[i]

		// stop the processing if matched volume has reached its max or orders can't filled
		if order.Price.GT(clearanceState.Price) || matchedVolume.Equal(clearanceState.MaxAskVolume) {
			break
		}

		// adjust the fill quantity using ProRata concept (proportionate bids/asks execution if demand/supply are not equal)
		fillQuantity := order.Quantity
		if proRataGTOne {
			orderQtyDec := sdk.NewDecFromBigInt(fillQuantity.BigInt())
			orderQtyDec = orderQtyDec.Mul(clearanceState.ProRataInvert)
			orderQtyDec = orderQtyDec.Ceil()

			if matchedVolume.Add(orderQtyDec).GT(clearanceState.MaxAskVolume) {
				orderQtyDec = clearanceState.MaxAskVolume.Sub(matchedVolume)
			}
			fillQuantity = sdk.NewUintFromBigInt(orderQtyDec.RoundInt().BigInt())
		}

		if fillQuantity.IsZero() {
			continue
		}
		matchedVolume = matchedVolume.Add(sdk.NewDecFromBigInt(fillQuantity.BigInt()))

		fills = append(fills, orderTypes.OrderFill{
			Order:            *order,
			ClearancePrice:   clearanceState.Price,
			QuantityFilled:   fillQuantity,
			QuantityUnfilled: order.Quantity.Sub(fillQuantity),
		})
	}

	return
}

// GetSDCurves returns SDCurves (for debug use only).
func (m *Matcher) GetSDCurves() SDCurves {
	return m.sdCurves
}

// NewMatcher creates a new Matcher object.
func NewMatcher(marketID dnTypes.ID, logger log.Logger) *Matcher {
	return &Matcher{
		marketID: marketID,
		logger:   logger,
	}
}
