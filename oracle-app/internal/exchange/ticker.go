package exchange

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	Code     string         `mapstructure:"code"`
	Pair     Pair           `mapstructure:"pair"`
	Simulate SimulateParams `mapstructure:"simulate"`
}

type SimulateParams struct {
	Enabled  bool `mapstructure:"enabled"`
	MinPrice int  `mapstructure:"min_price"`
	MaxPrice int  `mapstructure:"max_price"`
	PeriodS  int  `mapstructure:"period_s"`
}

func NewAsset(code string, pair Pair) Asset {
	return Asset{Code: code, Pair: pair}
}

type CurrencyPair = goex.CurrencyPair

type Ticker struct {
	Asset      Asset
	Price      sdk.Int
	Exchange   string
	ReceivedAt time.Time
}

func (t Ticker) String() string {
	return fmt.Sprintf("Asset: %s Price: %s ReceivedAt: %v", t.Asset.Code, t.Price, t.ReceivedAt)
}

func NewTicker(asset Asset, price sdk.Int, exchange string, receivedAt time.Time) Ticker {
	return Ticker{Asset: asset, Price: price, Exchange: exchange, ReceivedAt: receivedAt}
}
