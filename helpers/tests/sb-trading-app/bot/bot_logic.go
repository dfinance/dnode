package bot

import (
	"fmt"
	"math/rand"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/orders"
)

func (b *Bot) Start(wg *sync.WaitGroup, stopCh chan bool) {
	b.stopCh = stopCh
	b.logger.Info("starting")

	b.updateBalances()
	b.subscribeToOrderEvents()

	sellOrdersCount := b.generateSellOrders(
		b.baseBalance,
		b.cfg.MMakingMinPrice,
		b.cfg.MMakingMaxPrice,
		b.cfg.MMakingMinBaseVolume,
		b.cfg.MMakingInitOrders/2,
	)
	buyOrdersCount := b.generateBuyOrders(
		b.quoteBalance,
		b.cfg.MMakingMinPrice,
		b.cfg.MMakingMaxPrice,
		b.cfg.MMakingMinBaseVolume,
		b.cfg.MMakingInitOrders/2,
	)
	b.logger.Info(fmt.Sprintf("market making: initial orders: Sells / Buys: %d / %d", sellOrdersCount, buyOrdersCount))

	b.updateBalances()

	go func() {
		defer wg.Done()

		<-b.stopCh
		b.logger.Info("stopping")
	}()
}

func (b *Bot) onOrderCloseMarketMakeMaking(source string) {
	//posted, direction := b.newBalanceBasedOrder()
	//if posted {
	//	b.logger.Info("market making on %q: posted %s order", source, direction)
	//} else {
	//	b.logger.Info("market making on %q: skipped", source)
	//}
}

func (b *Bot) onMarketPriceChangeMarketMaking() {
	sellOrdersCount, buyOrdersCount, lowerPriceLimit, upperPriceLimit := b.newBalanceBasedOrders()
	lowerPriceDec := b.cfg.QuoteCurrency.UintToDec(lowerPriceLimit)
	upperPriceDec := b.cfg.QuoteCurrency.UintToDec(upperPriceLimit)
	if sellOrdersCount == 0 && buyOrdersCount == 0 {
		b.logger.Info(fmt.Sprintf("market making on %q: [%s:%s]: skipped", lowerPriceDec, upperPriceDec, "marketPrice change"))
	} else {
		b.logger.Info(fmt.Sprintf("market making on %q: [%s:%s]: Sells / Buys: %d / %d", "marketPrice change", lowerPriceDec, upperPriceDec, sellOrdersCount, buyOrdersCount))
	}
}

func (b *Bot) newBalanceBasedOrder() (posted bool, retDirection string) {
	var direction orders.Direction

	defer func() {
		if posted {
			if direction == orders.AskDirection {
				retDirection = "Sell"
			} else {
				retDirection = "Buy"
			}
		}
	}()

	var marketPrice sdk.Uint
	curBuyOrders, curSellOrders := 0, 0

	b.RLock()
	marketPrice = b.marketPrice
	for _, o := range b.orders {
		if o.Direction == orders.BidDirection {
			curBuyOrders++
		} else {
			curSellOrders++
		}
	}
	b.RUnlock()

	if marketPrice.IsZero() {
		return
	}

	if curBuyOrders > curSellOrders {
		direction = orders.AskDirection
	} else if curSellOrders > curBuyOrders {
		direction = orders.BidDirection
	} else {
		if coinDrop := rand.Uint32() % 2; coinDrop == 0 {
			direction = orders.AskDirection
		} else {
			direction = orders.BidDirection
		}
	}

	baseBalance, quoteBalance := b.updateBalances()

	switch direction {
	case orders.BidDirection:
		if quoteBalance.IsZero() {
			return
		}

		dampedPrice := b.dampPriceUp(b.marketPrice, b.marketPrice)
		if dampedPrice.Equal(b.lastPostedBidPrice) {
			dampedPrice = dampedPrice.Incr()
		}
		b.lastPostedBidPrice = dampedPrice

		quantity := quoteBalance.Quo(dampedPrice)

		posted = b.postBuyOrder(dampedPrice, quantity)
	case orders.AskDirection:
		if baseBalance.IsZero() {
			return
		}

		dampedPrice := b.dampPriceDown(b.marketPrice, b.marketPrice)
		if dampedPrice.Equal(b.lastPostedAskPrice) {
			dampedPrice = dampedPrice.Decr()
		}
		b.lastPostedAskPrice = dampedPrice

		posted = b.postBuyOrder(dampedPrice, baseBalance)
	}

	return
}

func (b *Bot) newBalanceBasedOrders() (sellOrdersCount, buyOrdersCount uint, priceLowerLimit, priceUpperLimit sdk.Uint) {
	var marketPrice sdk.Uint
	b.RLock()
	marketPrice = b.marketPrice
	b.RUnlock()

	if marketPrice.GT(b.cfg.MMakingMaxPrice) {
		priceUpperLimit = marketPrice
		priceLowerLimit = b.cfg.MMakingMaxPrice

	} else if marketPrice.LT(b.cfg.MMakingMinPrice) {
		priceUpperLimit = b.cfg.MMakingMinPrice
		priceLowerLimit = marketPrice
	} else {
		if rand.Uint64()%2 == 0 {
			priceUpperLimit = b.cfg.MMakingMaxPrice
			priceLowerLimit = marketPrice
		} else {
			priceUpperLimit = marketPrice
			priceLowerLimit = b.cfg.MMakingMinPrice
		}
	}

	if priceUpperLimit.Equal(priceLowerLimit) {
		return
	}

	baseBalance, quoteBalance := b.updateBalances()

	sellOrdersCount = b.generateSellOrders(
		baseBalance,
		priceLowerLimit,
		priceUpperLimit,
		b.cfg.MMakingMinBaseVolume,
		b.cfg.MMakingInitOrders/2,
	)

	buyOrdersCount = b.generateBuyOrders(
		quoteBalance,
		priceLowerLimit,
		priceUpperLimit,
		b.cfg.MMakingMinBaseVolume,
		b.cfg.MMakingInitOrders/2,
	)

	return
}

func (b *Bot) generateSellOrders(balance, minPrice, maxPrice, minVolume sdk.Uint, maxOrders uint64) uint {
	if balance.LT(minVolume) {
		return 0
	}

	ordersPrice, ordersQuantity := make([]sdk.Uint, 0), make([]sdk.Uint, 0)

	orderCount := maxOrders
	if count := balance.Quo(minVolume); count.LT(sdk.NewUint(maxOrders)) {
		orderCount = count.Uint64()
	}

	orderQuantity := balance.QuoUint64(orderCount)
	priceStep := maxPrice.Sub(minPrice).QuoUint64(orderCount)
	for i := uint64(0); i < orderCount; i++ {
		curStep := priceStep.MulUint64(i)
		price := minPrice.Add(curStep)
		priceWithNoise := b.dampPriceRandom(price, minPrice)

		if priceWithNoise.GTE(b.cfg.MMakingMinPrice) {
			ordersPrice = append(ordersPrice, priceWithNoise)
			ordersQuantity = append(ordersQuantity, orderQuantity)
		}
	}

	return b.postOrders(ordersPrice, ordersQuantity, orders.AskDirection)
}

func (b *Bot) generateBuyOrders(balance, minPrice, maxPrice, minVolume sdk.Uint, maxOrders uint64) uint {
	if balance.LT(minPrice) {
		return 0
	}

	minVolumeDec := b.cfg.BaseCurrency.UintToDec(minVolume)

	ordersPrice, ordersQuantity := make([]sdk.Uint, 0), make([]sdk.Uint, 0)

	orderCount := maxOrders
	if count := balance.Quo(minPrice); count.LT(sdk.NewUint(maxOrders)) {
		orderCount = count.Uint64()
	}

	amountStep := balance.QuoUint64(orderCount)
	amountStepDec := b.cfg.QuoteCurrency.UintToDec(amountStep)
	priceStep := maxPrice.Sub(minPrice).QuoUint64(orderCount)
	for i := uint64(0); i < orderCount; i++ {
		curStep := priceStep.MulUint64(i)
		price := minPrice.Add(curStep)
		priceWithNoise := b.dampPriceRandom(price, minPrice)
		priceWithNoiseDec := b.cfg.QuoteCurrency.UintToDec(priceWithNoise)

		quantityDec := amountStepDec.Quo(priceWithNoiseDec)
		if quantityDec.GTE(minVolumeDec) {
			ordersPrice = append(ordersPrice, priceWithNoise)
			ordersQuantity = append(ordersQuantity, b.cfg.BaseCurrency.DecToUint(quantityDec))
		}
	}

	return b.postOrders(ordersPrice, ordersQuantity, orders.BidDirection)
}
