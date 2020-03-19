package binance

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	goex "github.com/nntaoli-project/GoEx"
	ws "github.com/nntaoli-project/GoEx/binance"

	. "github.com/dfinance/dnode/oracle-app/internal/exchange"
	"github.com/dfinance/dnode/oracle-app/internal/utils"
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
	price, err := utils.FloatToFPString(t.Last, utils.Precision)
	if err != nil {
		fmt.Printf("error converting price: %v\n", err)
	}
	if old, found := e.lp.Load(t.Pair.String()); found && old.(string) == price {
		return
	} else {
		e.lp.Store(t.Pair.String(), price)
	}

	intPrice, isOk := sdk.NewIntFromString(price)
	if !isOk {
		fmt.Printf("error during parsing int price %q to bigint", price)
	}
	select {
	case out.(chan Ticker) <- NewTicker(
		NewAsset(fmt.Sprintf("%s_%s", strings.ToLower(pair.BaseAsset), strings.ToLower(pair.QuoteAsset)), pair),
		intPrice,
		exchangeName,
		ConvertTickerUnixMsTime(t.Date, time.Now().UTC(), 1*time.Hour)):
	default:
	}
}

// func (e *exchange) Name() string {
// 	return exchangeName
// }
