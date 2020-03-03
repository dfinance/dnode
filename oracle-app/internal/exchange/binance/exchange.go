package binance

import (
	"fmt"
	"sync"

	goex "github.com/nntaoli-project/GoEx"
	ws "github.com/nntaoli-project/GoEx/binance"

	. "github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange"
)

var _ Subscriber = (*exchange)(nil)

const exchangeName = "binance"

func init() {
	Register(exchangeName, New())
}

type exchange struct {
	ws   *ws.BinanceWs
	lp   *sync.Map // last price for assets
	subs *sync.Map // subscriptions out channels
}

func New() *exchange {
	e := exchange{ws: ws.NewBinanceWs(), lp: new(sync.Map), subs: new(sync.Map)}
	e.ws.SetCallbacks(e.tickerHandler, nil, nil, nil)

	return &e
}

func (e *exchange) Subscribe(a Asset, out chan Ticker) error {
	e.subs.Store(a.Pair.ID(), out)
	return e.ws.SubscribeTicker(CurrencyPair{
		CurrencyA: goex.Currency{Symbol: a.Pair.BaseAsset},
		CurrencyB: goex.Currency{Symbol: a.Pair.QuoteAsset},
	})
}

func (e *exchange) tickerHandler(t *goex.Ticker) {
	pair := NewPair(t.Pair.CurrencyA.Symbol, t.Pair.CurrencyB.Symbol)
	out, found := e.subs.Load(pair.ID())
	if !found {
		return
	}
	price := goex.FloatToString(t.Last, 8)
	if old, found := e.lp.Load(t.Pair.String()); found && old.(string) == price {
		return
	} else {
		e.lp.Store(t.Pair.String(), price)
	}
	select {
	case out.(chan Ticker) <- NewTicker(NewAsset(fmt.Sprintf("%s_%s", pair.BaseAsset, pair.QuoteAsset), pair), price, exchangeName):
	default:
	}
}

// func (e *exchange) Name() string {
// 	return exchangeName
// }
