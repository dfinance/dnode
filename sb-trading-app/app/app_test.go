package app

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/sb-trading-app/watcher"
)

func Test_SB_Trading(t *testing.T) {
	ct := cliTester.New(
		t,
		true,
		cliTester.LogLevel("main:error,state:info,x/orderbook:debug"),
		cliTester.DefaultConsensusTimings(),
	)
	defer ct.Close()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	watcherConfig := watcher.Config{
		T:             t,
		Tester:        ct,
		MinBots:       1,
		MaxBots:       0,
		WorkDurtInSec: 65,
		Markets: []watcher.Market{
			watcher.Market{
				BaseDenom:           "btc",
				QuoteDenom:          "dfi",
				InitMinPrice:        sdk.NewUint(10),
				InitMaxPrice:        sdk.NewUint(1000),
				InitOrders:          20,
				BaseSupply:          sdk.NewUint(1000 * 100000000),
				QuoteSupply:         sdk.NewUint(10000 * 100000000),
				OrderTtlInSec:       60,
				PriceDampingPercent: 15.0,
			},
		},
	}

	watcherObj := watcher.New(logger, watcherConfig)
	watcherObj.Work()
}
