package bot

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

func (b *Bot) updateBalances() {
	q, acc := b.cfg.Tester.QueryAccount(b.cfg.Address)
	require.NoError(b.cfg.T, executeQuery(q), "QueryAccount on updateBalances")

	b.Lock()
	defer b.Unlock()

	for _, coin := range acc.Coins {
		if coin.Denom == string(b.cfg.BaseCurrency.Denom) {
			b.baseBalance = sdk.Uint(coin.Amount)
			//b.logger.Info(fmt.Sprintf("baseBalance upd: %s", b.baseBalance))
			continue
		}
		if coin.Denom == string(b.cfg.QuoteCurrency.Denom) {
			b.quoteBalance = sdk.Uint(coin.Amount)
			//b.logger.Info(fmt.Sprintf("quoteBalance upd: %s", b.baseBalance))
			continue
		}
	}
}

func (b *Bot) postBuyOrder(price, quantity sdk.Uint) {
	b.postOrder(price, quantity, orderTypes.BidDirection)
}

func (b *Bot) postSellOrder(price, quantity sdk.Uint) {
	b.postOrder(price, quantity, orderTypes.AskDirection)
}

func (b *Bot) postOrder(price, quantity sdk.Uint, direction orderTypes.Direction) {
	txCmd := b.cfg.Tester.TxOrdersPost(b.cfg.Address, b.cfg.MarketID, direction, price, quantity, b.cfg.OrderTtlInSec)
	txCmd.DisableBroadcastMode()

	txCmd.SetAccountNumber(b.cfg.Number)
	txCmd.SetSequenceNumber(b.sequence)
	txHash := txCmd.CheckSucceeded()

	b.logger.Info(fmt.Sprintf("order posted (seq: %d) ([%s] %s -> %s): %s", b.sequence, direction, price, quantity, txHash))

	b.sequence++
}

func executeQuery(req *cliTester.QueryRequest) error {
	var lastErr error
	for i := 0; i < 15; i++ {
		output, err := req.Execute()
		if err == nil {
			return nil
		}

		lastErr = err
		if strings.Contains(output, "resource temporarily unavailable") {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		return err
	}

	return fmt.Errorf("retry failed: %v", lastErr)
}

func executeTx(req *cliTester.TxRequest) (string, error) {
	var lastErr error
	for i := 0; i < 10; i++ {
		txResp, err := req.Execute()
		if err == nil {
			return txResp.TxHash, nil
		}

		lastErr = err
		if strings.Contains(txResp.RawLog, "signature verification failed") {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		return "", err
	}

	return "", fmt.Errorf("retry failed: %v", lastErr)
}

func (b *Bot) subscribeToOrderEvents() {
	b.Lock()
	defer b.Unlock()

	// post events
	{
		query := fmt.Sprintf("orders.post.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		b.subs = append(b.subs, subscribeState{stopFunc: stopFunc})

		go func() {
			for {
				if event, ok := <-ch; ok {
					b.handleOrderPost(event)
				} else {
					return
				}
			}
		}()
	}

	// cancel events
	{
		query := fmt.Sprintf("orders.cancel.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		b.subs = append(b.subs, subscribeState{stopFunc: stopFunc})

		go func() {
			for {
				if event, ok := <-ch; ok {
					b.handleOrderCancel(event)
				} else {
					return
				}
			}
		}()
	}

	// fullyFilled events
	{
		query := fmt.Sprintf("orders.full_fill.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		b.subs = append(b.subs, subscribeState{stopFunc: stopFunc})

		go func() {
			for {
				if event, ok := <-ch; ok {
					b.handleOrderFullFill(event)
				} else {
					return
				}
			}
		}()
	}

	// partiallyFilled events
	{
		query := fmt.Sprintf("orders.partial_fill.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		b.subs = append(b.subs, subscribeState{stopFunc: stopFunc})

		go func() {
			for {
				if event, ok := <-ch; ok {
					b.handleOrderPartialFill(event)
				} else {
					return
				}
			}
		}()
	}

	// clearance events
	{
		query := fmt.Sprintf("orderbook.clearance.market_id='%s'", b.cfg.MarketID.String())
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		b.subs = append(b.subs, subscribeState{stopFunc: stopFunc})

		go func() {
			for {
				if event, ok := <-ch; ok {
					b.handleOrderBookClearance(event)
				} else {
					return
				}
			}
		}()
	}
}
