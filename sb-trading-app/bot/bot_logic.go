package bot

import (
	"fmt"
	"math/rand"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	orderTypes "github.com/dfinance/dnode/x/orders"
)

func (b *Bot) Start(wg *sync.WaitGroup, stopCh chan bool) {
	b.logger.Info("starting")

	b.updateBalances()
	b.subscribeToOrderEvents()

	// post initial orders with ranged rand price and maxed out quantity
	minPrice, maxPrice := b.cfg.InitMinPrice.Uint64(), b.cfg.InitMaxPrice.Uint64()
	diffPrice := maxPrice - minPrice
	buyOrders, sellOrders := b.cfg.InitOrders/2, b.cfg.InitOrders/2
	buyQuantity, sellQuantity := b.quoteBalance.QuoUint64(buyOrders).Uint64(), b.baseBalance.QuoUint64(sellOrders).Uint64()

	for i := uint64(0); i < sellOrders; i++ {
		price := minPrice + rand.Uint64()%(diffPrice+1)
		b.postSellOrder(sdk.NewUint(price), sdk.NewUint(sellQuantity))
	}
	for i := uint64(0); i < buyOrders; i++ {
		price := minPrice + rand.Uint64()%(diffPrice+1)
		b.postBuyOrder(sdk.NewUint(price), sdk.NewUint(buyQuantity))
	}

	b.updateBalances()

	go func() {
		defer wg.Done()

		<-stopCh
		b.close()
	}()
}

func (b *Bot) newOrder() {
	curBuyOrders, curSellOrders := 0, 0
	var marketPrice sdk.Uint
	var baseBalance, quoteBalance sdk.Uint

	b.RLock()
	marketPrice = b.marketPrice
	baseBalance = b.baseBalance
	quoteBalance = b.quoteBalance
	for _, o := range b.orders {
		if o.Direction == orderTypes.BidDirection {
			curBuyOrders++
		} else {
			curSellOrders++
		}
	}
	b.RUnlock()

	if marketPrice.IsZero() {
		return
	}

	var newDirection orderTypes.Direction
	if curBuyOrders > curSellOrders {
		newDirection = orderTypes.AskDirection
	} else if curSellOrders > curBuyOrders {
		newDirection = orderTypes.BidDirection
	} else {
		if coinDrop := rand.Uint32() % 2; coinDrop == 0 {
			newDirection = orderTypes.AskDirection
		} else {
			newDirection = orderTypes.BidDirection
		}
	}

	b.updateBalances()

	marketPriceFloat := float64(marketPrice.Uint64())
	switch newDirection {
	case orderTypes.BidDirection:
		if quoteBalance.IsZero() {
			return
		}

		dampedPriceFloat := marketPriceFloat + marketPriceFloat / 100.0 * b.cfg.NewOrderDampingPercent
		dampedPrice := sdk.NewUint(uint64(dampedPriceFloat))
		if dampedPrice.Equal(b.lastPostedBidPrice) {
			dampedPrice = dampedPrice.Incr()
		}
		b.lastPostedBidPrice = dampedPrice

		quantity := quoteBalance.Quo(dampedPrice)

		b.postBuyOrder(dampedPrice, quantity)
	case orderTypes.AskDirection:
		if baseBalance.IsZero() {
			return
		}

		dampedPriceFloat := marketPriceFloat - marketPriceFloat / 100.0 * b.cfg.NewOrderDampingPercent
		dampedPrice := sdk.NewUint(uint64(dampedPriceFloat))
		if dampedPrice.Equal(b.lastPostedAskPrice) {
			dampedPrice = dampedPrice.Decr()
		}
		b.lastPostedAskPrice = dampedPrice

		b.postBuyOrder(dampedPrice, baseBalance)
	}

	b.logger.Info(fmt.Sprintf("new order (%s): posted", newDirection))
}
