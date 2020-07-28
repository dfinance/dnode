// +build debug

package app

import (
	"os"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/helpers/tests/sb-trading-app/watcher"
)

const (
	DecimalsDFI = "1000000000000000000"
	DecimalsBTC = "100000000"
)

func Test_SB_Trading(t *testing.T) {
	const (
		minClientsPerMarket = 1
		maxClientsPerMarket = 4
		workDurInSec        = 300
		initOrders          = 50
	)

	oneDfi := sdk.NewUintFromString(DecimalsDFI)
	oneBtc := sdk.NewUintFromString(DecimalsBTC)
	markets := []watcher.Market{
		watcher.Market{
			BaseDenom:            "btc",
			QuoteDenom:           "dfi",
			OrderTtlInSec:        60,
			MMakingMinPrice:      sdk.NewUint(10).Mul(oneDfi),
			MMakingMaxPrice:      sdk.NewUint(10000).Mul(oneDfi),
			MMakingMinBaseVolume: sdk.NewUint(1).Mul(oneBtc),
			BaseSupply:           sdk.NewUint(10000).Mul(oneBtc),
			QuoteSupply:          sdk.NewUint(100000000).Mul(oneDfi),
			MMakingInitOrders:    initOrders,
			PriceDampingPercent:  5,
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
		cliTester.DaemonLogLevelOption("main:error,state:info,x/orders:error,x/orderbook:info"),
		cliTester.AccountsOption(accountOpts...),
		//cliTester.DefaultConsensusTimingsOption(),
		cliTester.ConsensusTimingsOption(
			"3s",
			"500ms",
			"1s",
			"500ms",
			"1s",
			"500ms",
			"10s",
		),
		//cliTester.MempoolOption(500000, 1000000, 104857600, 107374182400),
	)
	ct.StartRestServer(false)
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
