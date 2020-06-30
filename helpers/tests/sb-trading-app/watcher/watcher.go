package watcher

import (
	"fmt"
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/helpers/tests/sb-trading-app/bot"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	ccTypes "github.com/dfinance/dnode/x/currencies"
)

type Watcher struct {
	logger       log.Logger
	cfg          Config
	marketStates []*MarketState
	history      *History
	curBots      uint
	wg           *sync.WaitGroup
	stopCh       chan bool
}

type Config struct {
	T             *testing.T
	Tester        *cliTester.CLITester
	MinBots       uint
	MaxBots       uint
	WorkDurtInSec int
	Markets       []Market
}

type Market struct {
	BaseDenom            string
	QuoteDenom           string
	BaseSupply           sdk.Uint
	QuoteSupply          sdk.Uint
	OrderTtlInSec        int
	PriceDampingPercent  uint64
	MMakingMinPrice      sdk.Uint
	MMakingMaxPrice      sdk.Uint
	MMakingMinBaseVolume sdk.Uint
	MMakingInitOrders    uint64
}

type MarketState struct {
	Market
	id            dnTypes.ID
	baseCurrency  ccTypes.Currency
	quoteCurrency ccTypes.Currency
	bots          []*bot.Bot
}

func New(logger log.Logger, cfg Config) *Watcher {
	w := &Watcher{
		logger: logger.With("watcher", ""),
		cfg:    cfg,
		wg:     &sync.WaitGroup{},
		stopCh: make(chan bool),
	}

	q, _ := w.cfg.Tester.QueryStatus()
	q.RemoveCmdArg("status")
	w.logger.Info(q.String())

	marketCreator := w.cfg.Tester.Accounts["validator1"].Address
	marketInfos := make(map[string]MarketInfo, len(cfg.Markets))
	for _, marketCfg := range cfg.Markets {
		marketState := &MarketState{Market: marketCfg}

		w.logger.Info(fmt.Sprintf("market init: %s / %s", marketState.BaseDenom, marketState.QuoteDenom))
		w.cfg.Tester.TxMarketsAdd(marketCreator, marketState.BaseDenom, marketState.QuoteDenom).CheckSucceeded()

		q, markets := w.cfg.Tester.QueryMarketsList(-1, -1, &marketState.BaseDenom, &marketState.QuoteDenom)
		q.CheckSucceeded()
		require.Len(w.cfg.T, *markets, 1, "market not created")
		marketState.id = (*markets)[0].ID

		q, baseInfo := w.cfg.Tester.QueryCurrenciesCurrency(marketState.BaseDenom)
		q.CheckSucceeded()
		marketState.baseCurrency = *baseInfo

		q, quoteInfo := w.cfg.Tester.QueryCurrenciesCurrency(marketState.QuoteDenom)
		q.CheckSucceeded()
		marketState.quoteCurrency = *quoteInfo

		for i := uint(0); i < w.cfg.MaxBots; i++ {
			clientName := NewClientName(int(i), marketState.Market)
			account, ok := w.cfg.Tester.Accounts[clientName]
			require.True(w.cfg.T, ok, "account not found in CLITester: %s", clientName)

			botObj := bot.New(logger, bot.Config{
				T:                      w.cfg.T,
				Tester:                 w.cfg.Tester,
				Name:                   clientName,
				Address:                account.Address,
				Number:                 account.Number,
				BaseCurrency:           marketState.baseCurrency,
				QuoteCurrency:          marketState.quoteCurrency,
				MarketID:               marketState.id,
				MMakingMinPrice:        marketState.MMakingMinPrice,
				MMakingMaxPrice:        marketState.MMakingMaxPrice,
				MMakingInitOrders:      marketState.MMakingInitOrders,
				MMakingMinBaseVolume:   marketState.MMakingMinBaseVolume,
				OrderTtlInSec:          marketState.OrderTtlInSec,
				NewOrderDampingPercent: marketState.PriceDampingPercent,
			})

			marketState.bots = append(marketState.bots, botObj)
		}

		w.marketStates = append(w.marketStates, marketState)
		marketInfos[marketState.id.String()] = MarketInfo{
			BaseCurrency:  marketState.baseCurrency,
			QuoteCurrency: marketState.quoteCurrency,
		}
	}

	w.curBots = w.cfg.MinBots
	w.history = NewHistory(w.cfg.T, marketInfos, w.curBots)

	return w
}

func NewClientName(clientIdx int, marketCfg Market) string {
	return fmt.Sprintf("%s_%s_client_%d", marketCfg.BaseDenom, marketCfg.QuoteDenom, clientIdx)
}
