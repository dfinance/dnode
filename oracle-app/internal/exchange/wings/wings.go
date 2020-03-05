package wings

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/WingsDao/wings-blockchain/oracle-app/internal/exchange"
)

const (
	exchangeName = "wings"

	basePriceMin = 230
	basePriceMax = 250
)

var _ Subscriber = (*wings)(nil)

type wings struct{}

func init() {
	Register(exchangeName, &wings{})
}

func (w wings) Subscribe(_ Asset, out chan Ticker) error {
	rand.Seed(time.Now().UnixNano())
	ticker := time.Tick(time.Second)
	go func() {
		for {
			<-ticker
			randPrice := rand.Intn(basePriceMax-basePriceMin) + basePriceMin
			priceWingsEth := fmt.Sprintf("%.8f", float64(randPrice)/1000)
			priceEthWings := fmt.Sprintf("%.8f", float64(randPrice))
			out <- NewTicker(NewAsset("wings_eth", Pair{}), priceWingsEth, "wings-test")
			out <- NewTicker(NewAsset("eth_wings", Pair{}), priceEthWings, "wings-test")
		}
	}()
	return nil
}
