package dfinance

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/dfinance/dnode/oracle-app/internal/exchange"
)

const (
	exchangeName = "dfinance"

	basePriceMin = 230
	basePriceMax = 250
)

var _ Subscriber = (*dnSubscriber)(nil)

type dnSubscriber struct{}

func init() {
	Register(exchangeName, &dnSubscriber{})
}

func (d dnSubscriber) Subscribe(_ Asset, out chan Ticker) error {
	rand.Seed(time.Now().UnixNano())
	ticker := time.Tick(time.Second)
	go func() {
		for {
			<-ticker
			randPrice := rand.Intn(basePriceMax-basePriceMin) + basePriceMin
			priceDfiEth := fmt.Sprintf("%.8f", float64(randPrice)/1000)
			priceEthDfi := fmt.Sprintf("%.8f", float64(randPrice))
			out <- NewTicker(NewAsset("dfi_eth", Pair{}), priceDfiEth, "dfi-test", time.Now().UTC())
			out <- NewTicker(NewAsset("eth_dfi", Pair{}), priceEthDfi, "dfi-test", time.Now().UTC())
		}
	}()
	return nil
}
