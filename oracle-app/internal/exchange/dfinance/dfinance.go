package dfinance

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/types"

	. "github.com/dfinance/dnode/oracle-app/internal/exchange"
)

const (
	exchangeName = "dfinance-test"
)

var _ Subscriber = (*dnSubscriber)(nil)

type dnSubscriber struct{}

func init() {
	Register(exchangeName, &dnSubscriber{})
}

func (d dnSubscriber) Subscribe(asset Asset, out chan Ticker) error {
	if !asset.Simulate.Enabled {
		return fmt.Errorf("asset %s: simulation disabled", asset.Code)
	}
	if asset.Simulate.PeriodS <= 0 {
		return fmt.Errorf("asset %s: invalid simulation period", asset.Code)
	}
	if asset.Simulate.MinPrice < 0 || asset.Simulate.MaxPrice < 0 {
		return fmt.Errorf("asset %s: invalid simulation minPrice/maxPrice: lt 0", asset.Code)
	}
	if asset.Simulate.MinPrice >= asset.Simulate.MaxPrice {
		return fmt.Errorf("asset %s: invalid simulation minPrice/maxPrice: min ge max", asset.Code)
	}

	rand.Seed(time.Now().UnixNano())
	ticker := time.Tick(time.Duration(asset.Simulate.PeriodS) * time.Second)
	go func() {
		for {
			<-ticker

			randPrice := rand.Intn(asset.Simulate.MaxPrice-asset.Simulate.MinPrice) + asset.Simulate.MinPrice

			out <- NewTicker(NewAsset(asset.Code, Pair{}), types.NewInt(int64(randPrice)), exchangeName, time.Now().UTC())
		}
	}()

	return nil
}
