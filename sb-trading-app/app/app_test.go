package app

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/sb-trading-app/watcher"
)

func Test_SB_Trading(t *testing.T) {
	const (
		minClientsPerMarket = 10
		maxClientsPerMarket = 10
		workDurInSec        = 200
	)

	markets := []watcher.Market{
		watcher.Market{
			BaseDenom:            "btc",
			QuoteDenom:           "dfi",
			OrderTtlInSec:        60,
			BaseSupply:           sdk.NewUint(1000 * 100000000),
			QuoteSupply:          sdk.NewUint(10000 * 100000000),
			MMakingMinPrice:      sdk.NewUint(10),
			MMakingMaxPrice:      sdk.NewUint(1000),
			MMakingInitOrders:    20,
			MMakingMinBaseVolume: 10,
			PriceDampingPercent:  5.0,
		},
	}

	maxAccounts := maxClientsPerMarket * len(markets)
	accountOpts := make([]cliTester.AccountOption, maxAccounts, maxAccounts)
	for marketIdx := 0; marketIdx < len(markets); marketIdx++ {
		market := markets[marketIdx]
		for clientIdx := 0; clientIdx < maxClientsPerMarket; clientIdx++ {
			account := &accountOpts[marketIdx*maxClientsPerMarket+clientIdx]
			account.Name = watcher.NewClientName(clientIdx, market)
			account.Balances = []cliTester.StringPair{
				{
					Key:   market.BaseDenom,
					Value: market.BaseSupply.String(),
				},
				{
					Key:   market.QuoteDenom,
					Value: market.QuoteSupply.String(),
				},
			}
		}
	}

	ct := cliTester.New(
		t,
		true,
		cliTester.LogLevel("main:error,state:error,x/orders:error,x/orderbook:error"),
		cliTester.DefaultConsensusTimings(),
		cliTester.Accounts(accountOpts...),

	)
	defer ct.Close()

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	filteredLogger, err := tmFlags.ParseLogLevel("watcher:info,client:info", logger, "info")
	require.NoError(t, err, "logLevel option")

	watcherConfig := watcher.Config{
		T:             t,
		Tester:        ct,
		MinBots:       minClientsPerMarket,
		MaxBots:       maxClientsPerMarket,
		WorkDurtInSec: workDurInSec,
		Markets:       markets,
	}

	watcherObj := watcher.New(filteredLogger, watcherConfig)
	watcherObj.Work()
}
