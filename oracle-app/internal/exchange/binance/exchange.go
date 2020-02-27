package binance

import (
	"sync"

	goex "github.com/nntaoli-project/GoEx"
	ws "github.com/nntaoli-project/GoEx/binance"
	"github.com/sirupsen/logrus"

	. "github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange"
)

var _ Subscriber = (*Exchange)(nil)

type Exchange struct {
	log  *logrus.Logger
	ws   *ws.BinanceWs
	lp   *sync.Map // last price for assets
	subs *sync.Map // subscriptions out channels
}

func New(l *logrus.Logger) *Exchange {
	e := Exchange{ws: ws.NewBinanceWs(), lp: new(sync.Map), subs: new(sync.Map), log: l}
	e.ws.SetCallbacks(e.tickerHandler, nil, nil, nil)

	return &e
}

func (e *Exchange) Subscribe(a Asset, out chan Ticker) error {
	e.subs.Store(a.Pair.ID(), out)
	return e.ws.SubscribeTicker(CurrencyPair{
		CurrencyA: goex.Currency{Symbol: a.Pair.BaseAsset},
		CurrencyB: goex.Currency{Symbol: a.Pair.QuoteAsset},
	})
}

func (e *Exchange) tickerHandler(t *goex.Ticker) {
	e.log.Debug(t)
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
	case out.(chan Ticker) <- NewTicker(NewAsset(pair.ID(), pair), price):
	default:
	}
}
