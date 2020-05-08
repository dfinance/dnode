package keeper

import (
	"fmt"
	"sort"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	orderTypes "github.com/dfinance/dnode/x/order"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

type Matcher struct {
	marketID   dnTypes.ID
	logger     log.Logger
	orders     MatcherOrders
	aggregates MatcherAggregates
}

type MatcherOrders struct {
	bid orderTypes.Orders
	ask orderTypes.Orders
}

type MatcherAggregates struct {
	bid OrderAggregates
	ask OrderAggregates
}

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

	// =)
	if len(*orders) == MaxInt {
		return fmt.Errorf("max orders len reached")
	}

	*orders = append(*orders, *order)

	return nil
}

func (m *Matcher) Match() (result types.MatcherResult, retErr error) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		sort.Sort(ByPriceAscIDDesc(m.orders.bid))
		m.aggregates.bid = NewBidOrderAggregates(m.orders.bid)
	}()

	go func() {
		defer wg.Done()
		sort.Sort(ByPriceAscIDAsc(m.orders.ask))
		m.aggregates.ask = NewAskOrderAggregates(m.orders.ask)
	}()

	wg.Wait()

	pqCurves, err := NewPQCurves(m.aggregates.ask, m.aggregates.bid)
	if err != nil {
		retErr = err
		return
	}

	clearanceState, err := pqCurves.GetClearanceState()
	if err != nil {
		retErr = err
		return
	}

	bidFills, bidMatchedVolume := m.getBidOrderFills(clearanceState)
	askFills, askMatchedVolume := m.getAskOrderFills(clearanceState)

	result = types.MatcherResult{
		ClearanceState:   clearanceState,
		MatchedBidVolume: bidMatchedVolume,
		MatchedAskVolume: askMatchedVolume,
		OrderFills:       append(bidFills, askFills...),
	}

	m.logger.Debug(fmt.Sprintf("Matcher results for marketID %s:", m.marketID.String()))
	m.logger.Debug("Bid orders:")
	m.logger.Debug("\n" + m.orders.bid.String())
	m.logger.Debug("Bid aggregates:")
	m.logger.Debug("\n" + m.aggregates.bid.String())
	m.logger.Debug("Ask orders:")
	m.logger.Debug("\n" + m.orders.ask.String())
	m.logger.Debug("Ask aggregates:")
	m.logger.Debug("\n" + m.aggregates.ask.String())
	m.logger.Debug("PQ curves:")
	m.logger.Debug("\n" + pqCurves.String())
	m.logger.Debug("\n" + result.String())

	return
}

func (m *Matcher) getBidOrderFills(clearanceState types.ClearanceState) (fills orderTypes.OrderFills, matchedVolume sdk.Dec) {
	fills, matchedVolume = make(orderTypes.OrderFills, 0, len(m.orders.bid)), sdk.ZeroDec()

	proRataGTOne := clearanceState.ProRata.GT(sdk.OneDec())
	for i := len(m.orders.bid) - 1; i >= 0; i-- {
		order := &m.orders.bid[i]

		if order.Price.LT(clearanceState.Price) || matchedVolume.Equal(clearanceState.MaxBidVolume) {
			break
		}

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

func (m *Matcher) getAskOrderFills(clearanceState types.ClearanceState) (fills orderTypes.OrderFills, matchedVolume sdk.Dec) {
	fills, matchedVolume = make(orderTypes.OrderFills, 0, len(m.orders.ask)), sdk.ZeroDec()

	proRataGTOne := clearanceState.ProRata.GT(sdk.OneDec())
	for i := 0; i < len(m.orders.ask); i++ {
		order := &m.orders.ask[i]

		if order.Price.GT(clearanceState.Price) || matchedVolume.Equal(clearanceState.MaxAskVolume) {
			break
		}

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

func NewMatcher(marketID dnTypes.ID, logger log.Logger) *Matcher {
	return &Matcher{
		marketID: marketID,
		logger:   logger,
	}
}
