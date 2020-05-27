package bot

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/sb-trading-app/utils"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

type Api interface {
	GetAccount() (sequence uint64, baseBalance, quoteBalance sdk.Uint, retErr error)
	PostOrder(price, quantity sdk.Uint, direction orderTypes.Direction) (txHash string, retErr error)
	GetOrder(id dnTypes.ID) (order *orderTypes.Order, retErr error)
}

func (b *Bot) updateBalances() (baseBalance, quoteBalance sdk.Uint) {
	var err error
	_, baseBalance, quoteBalance, err = b.api.GetAccount()
	require.NoError(b.cfg.T, err)

	b.Lock()
	defer b.Unlock()

	b.baseBalance = baseBalance
	b.quoteBalance = quoteBalance

	return
}

func (b *Bot) postBuyOrder(price, quantity sdk.Uint) bool {
	return b.postOrder(price, quantity, orderTypes.BidDirection)
}

func (b *Bot) postSellOrder(price, quantity sdk.Uint) bool {
	return b.postOrder(price, quantity, orderTypes.AskDirection)
}

func (b *Bot) postOrders(price, quantity []sdk.Uint, direction orderTypes.Direction) uint {
	count := uint(0)
	for i := uint(0); i < uint(len(price)); i++ {
		if b.postOrder(price[i], quantity[i], direction) {
			count++
		}
	}
	return count
}

func (b *Bot) postOrder(price, quantity sdk.Uint, direction orderTypes.Direction) (posted bool) {
	txHash, err := b.api.PostOrder(price, quantity, direction)
	require.NoError(b.cfg.T, err)

	if txHash != "" {
		b.logger.Debug(fmt.Sprintf("order posted ([%s] %s -> %s): %s",
			direction,
			b.cfg.QuoteCurrency.UintToDec(price),
			b.cfg.BaseCurrency.UintToDec(quantity),
			txHash,
		))
		posted = true
	}

	return
}

func (b *Bot) subscribeToOrderEvents() {
	b.Lock()
	defer b.Unlock()

	// post events
	//go commonHandler(fmt.Sprintf("orders.post.owner='%s'", b.cfg.Address), b.handleOrderPost)

	// cancel events
	//go commonHandler(fmt.Sprintf("orders.cancel.owner='%s'", b.cfg.Address), b.handleOrderCancel)

	// fullyFilled events
	//go commonHandler(fmt.Sprintf("orders.full_fill.owner='%s'", b.cfg.Address), b.handleOrderFullFill)

	// partiallyFilled events
	//go commonHandler(fmt.Sprintf("orders.partial_fill.owner='%s'", b.cfg.Address), b.handleOrderPartialFill)

	// clearance events
	go utils.EventsWorker(b.logger, b.cfg.Tester, b.stopCh,
		fmt.Sprintf("orderbook.clearance.market_id='%s'", b.cfg.MarketID.String()),
		b.handleOrderBookClearance,
	)
}
