package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/logger"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	marketTypes "github.com/dfinance/dnode/x/market"
	orderTypes "github.com/dfinance/dnode/x/order"
)

func Test(t *testing.T) {
	inputs := []struct {
		Direction orderTypes.Direction
		ID       uint64
		Price    uint64
		Quantity uint64
	}{
		{orderTypes.BidDirection, 5, 12, 100},
		{orderTypes.AskDirection, 6, 10, 50},
		{orderTypes.AskDirection, 4, 10, 50},
		{orderTypes.BidDirection, 3, 12, 100},
		{orderTypes.AskDirection, 2, 8, 100},
		{orderTypes.AskDirection, 10, 14, 100},
		{orderTypes.BidDirection, 7, 14, 50},
		{orderTypes.AskDirection, 8, 11, 100},
		{orderTypes.BidDirection, 1, 10, 100},
	}

	market := marketTypes.NewMarket(dnTypes.NewIDFromUint64(0), "baseDenom", "quoteDenom", 0)

	testLogger := logger.NewDNLogger()
	testLogger = log.NewFilter(testLogger, log.AllowAll())
	matcherPool := NewMatcherPool(testLogger)

	for _, i := range inputs {
		order := orderTypes.Order{
			ID:        dnTypes.NewIDFromUint64(i.ID),
			Owner:     nil,
			Market:    market,
			Direction: i.Direction,
			Price:     sdk.NewUint(i.Price),
			Quantity:  sdk.NewUint(i.Quantity),
			Ttl:       time.Duration(i.Price),
			CreatedAt: time.Time{},
		}

		if err := matcherPool.AddOrder(order); err != nil {
			t.Fatalf("AddOrder: %v", err)
		}
	}

	matcherPool.Process()
}
