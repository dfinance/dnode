package bot

import (
	"sync"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/tests/clitester"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/orders"
)

type Bot struct {
	sync.RWMutex
	logger             log.Logger
	cfg                Config
	baseBalance        sdk.Uint
	quoteBalance       sdk.Uint
	marketPrice        sdk.Uint
	orders             map[string]orders.Order
	lastPostedAskPrice sdk.Uint
	lastPostedBidPrice sdk.Uint
	api                Api
	stopCh             chan bool
}

type Config struct {
	T                      *testing.T
	Tester                 *clitester.CLITester
	Name                   string
	Address                string
	Number                 uint64
	BaseCurrency           ccstorage.Currency
	QuoteCurrency          ccstorage.Currency
	MarketID               dnTypes.ID
	MMakingMinPrice        sdk.Uint
	MMakingMaxPrice        sdk.Uint
	MMakingInitOrders      uint64
	MMakingMinBaseVolume   sdk.Uint
	OrderTtlInSec          int
	NewOrderDampingPercent uint64
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

func New(logger log.Logger, cfg Config) *Bot {
	b := &Bot{
		logger:             logger.With("client", cfg.Name, "address", cfg.Address),
		cfg:                cfg,
		baseBalance:        sdk.ZeroUint(),
		quoteBalance:       sdk.ZeroUint(),
		marketPrice:        sdk.ZeroUint(),
		orders:             make(map[string]orders.Order),
		lastPostedAskPrice: sdk.ZeroUint(),
		lastPostedBidPrice: sdk.ZeroUint(),
	}

	//b.api = NewApiCli(b.cfg.Tester, b.cfg.Number, b.cfg.Address, b.cfg.MarketID, string(b.cfg.BaseCurrency.Denom), string(b.cfg.QuoteCurrency.Denom), b.cfg.OrderTtlInSec)
	b.api = NewApiRest(b.cfg.Tester, b.cfg.Number, b.cfg.Name, b.cfg.Address, b.cfg.MarketID, string(b.cfg.BaseCurrency.Denom), string(b.cfg.QuoteCurrency.Denom), b.cfg.OrderTtlInSec)

	return b
}
