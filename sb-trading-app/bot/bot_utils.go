package bot

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (b *Bot) randomDampingPercent() float64 {
	return float64(rand.Int63n(int64(b.cfg.NewOrderDampingPercent*100.0) + 1)) / 100.0
}

func (b *Bot) percentOfPrice(price sdk.Uint, percent float64) sdk.Uint {
	priceFloat := float64(price.Uint64())
	resultFloat := priceFloat / 100.0 * percent

	return sdk.NewUint(uint64(resultFloat))
}

func (b *Bot) dampPriceUp(price sdk.Uint) sdk.Uint {
	return price.Add(b.percentOfPrice(price, b.randomDampingPercent()))
}

func (b *Bot) dampPriceDown(price sdk.Uint) sdk.Uint {
	return price.Sub(b.percentOfPrice(price, b.randomDampingPercent()))
}

func(b *Bot) dampPriceRandom(price sdk.Uint) sdk.Uint {
	if rand.Uint64() % 2 == 0 {
		return b.dampPriceUp(price)
	}

	return b.dampPriceDown(price)
}
