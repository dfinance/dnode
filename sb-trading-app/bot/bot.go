package bot

import (
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	crTypes "github.com/dfinance/dnode/x/currencies_register"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

type Bot struct {
	sync.RWMutex
	logger             log.Logger
	cfg                Config
	baseBalance        sdk.Uint
	quoteBalance       sdk.Uint
	marketPrice        sdk.Uint
	orders             map[string]orderTypes.Order
	subs               []subscribeState
	blockHeight        int64
	sequence           uint64
	lastPostedAskPrice sdk.Uint
	lastPostedBidPrice sdk.Uint
}

type subscribeState struct {
	stopFunc func()
}

type Config struct {
	T                      *testing.T
	Tester                 *cliTester.CLITester
	Name                   string
	Address                string
	Number                 uint64
	BaseCurrency           crTypes.CurrencyInfo
	QuoteCurrency          crTypes.CurrencyInfo
	MarketID               dnTypes.ID
	InitMinPrice           sdk.Uint
	InitMaxPrice           sdk.Uint
	InitOrders             uint64
	OrderTtlInSec          int
	NewOrderDampingPercent float64
}

func (b *Bot) Name() string {
	return b.cfg.Name
}

func (b *Bot) Balances() (baseBalance, quoteBalance sdk.Uint) {
	b.RLock()
	defer b.RUnlock()

	baseBalance = b.baseBalance
	quoteBalance = b.quoteBalance

	return
}

func (b *Bot) close() {
	b.Lock()
	defer b.Unlock()

	for _, f := range b.subs {
		f.stopFunc()
	}
}

func New(logger log.Logger, cfg Config) *Bot {
	return &Bot{
		logger:             logger.With("client", cfg.Name, "address", cfg.Address),
		cfg:                cfg,
		baseBalance:        sdk.ZeroUint(),
		quoteBalance:       sdk.ZeroUint(),
		marketPrice:        sdk.ZeroUint(),
		orders:             make(map[string]orderTypes.Order, 0),
		lastPostedAskPrice: sdk.ZeroUint(),
		lastPostedBidPrice: sdk.ZeroUint(),
	}
}
