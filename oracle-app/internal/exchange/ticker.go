package exchange

import (
	"fmt"

	goex "github.com/nntaoli-project/GoEx"
)

type Pair struct {
	BaseAsset  string `mapstructure:"base"`
	QuoteAsset string `mapstructure:"quote"`
}

func NewPair(baseAsset string, quoteAsset string) Pair {
	return Pair{BaseAsset: baseAsset, QuoteAsset: quoteAsset}
}

func (p *Pair) CurrencyPair() CurrencyPair {
	return CurrencyPair{
		CurrencyA: goex.Currency{Symbol: p.BaseAsset},
		CurrencyB: goex.Currency{Symbol: p.QuoteAsset},
	}
}

func (p *Pair) ID() string {
	return p.BaseAsset + p.QuoteAsset
}

type Asset struct {
	Code string `mapstructure:"code"`
	Pair Pair   `mapstructure:"pair"`
}

func NewAsset(code string, pair Pair) Asset {
	return Asset{Code: code, Pair: pair}
}

type CurrencyPair = goex.CurrencyPair

type Ticker struct {
	Asset Asset
	Price string
}

func (t Ticker) String() string {
	return fmt.Sprintf("Asset: %s Price: %s", t.Asset.Code, t.Price)
}

func NewTicker(asset Asset, price string) Ticker {
	return Ticker{Asset: asset, Price: price}
}
