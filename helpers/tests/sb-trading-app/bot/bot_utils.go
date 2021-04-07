package bot

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (b *Bot) randomDampingPercent() uint64 {
	return rand.Uint64() % (b.cfg.NewOrderDampingPercent + 1)
}

func (b *Bot) percentOfPrice(price sdk.Uint, percent uint64) sdk.Uint {
	return price.QuoUint64(100).MulUint64(percent)
}

func (b *Bot) dampPriceUp(price, randomBase sdk.Uint) sdk.Uint {
	return price.Add(b.percentOfPrice(randomBase, b.randomDampingPercent()))
}

func (b *Bot) dampPriceDown(price, randomBase sdk.Uint) sdk.Uint {
	return price.Sub(b.percentOfPrice(randomBase, b.randomDampingPercent()))
}

func (b *Bot) dampPriceRandom(price, randomBase sdk.Uint) sdk.Uint {
	if rand.Uint64()%2 == 0 {
		return b.dampPriceUp(price, randomBase)
	}

	return b.dampPriceDown(price, randomBase)
}
