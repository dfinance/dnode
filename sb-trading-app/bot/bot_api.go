package bot

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

func (b *Bot) updateBalances() (baseBalance, quoteBalance sdk.Uint) {
	q, acc := b.cfg.Tester.QueryAccount(b.cfg.Address)
	require.NoError(b.cfg.T, executeQuery(q), "QueryAccount on updateBalances")

	b.Lock()
	defer b.Unlock()

	for _, coin := range acc.Coins {
		if coin.Denom == string(b.cfg.BaseCurrency.Denom) {
			b.baseBalance = sdk.Uint(coin.Amount)
			continue
		}
		if coin.Denom == string(b.cfg.QuoteCurrency.Denom) {
			b.quoteBalance = sdk.Uint(coin.Amount)
			continue
		}
	}
	baseBalance, quoteBalance = b.baseBalance, b.quoteBalance

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
	txHash := ""
	defer func() {
		if txHash != "" {
			b.logger.Debug(fmt.Sprintf("order posted (seq: %d) ([%s] %s -> %s): %s", b.sequence, direction, price, quantity, txHash))
			b.sequence++
			posted = true
		}
	}()

	if price.IsZero() || quantity.IsZero() {
		return
	}

	txCmd := b.cfg.Tester.TxOrdersPost(b.cfg.Address, b.cfg.MarketID, direction, price, quantity, b.cfg.OrderTtlInSec)
	txCmd.DisableBroadcastMode()
	txCmd.SetAccountNumber(b.cfg.Number)
	txCmd.SetSequenceNumber(b.sequence)
	resp, err := txCmd.Execute()
	if err == nil {
		txHash = resp.TxHash
		return
	}
	if resp.Code == 19 {
		return
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		require.NoError(b.cfg.T, err, "PostOrder: first attempt")
	}

	qAcc, acc := b.cfg.Tester.QueryAccount(b.cfg.Address)
	require.NoError(b.cfg.T, executeQuery(qAcc), "PostOrder: GetAccount")

	b.sequence = acc.Sequence
	var lastErr error
	for i := 0; i < 10; i++ {
		txCmd.RemoveCmdArg("sequence")
		txCmd.SetSequenceNumber(b.sequence)
		resp, err := txCmd.Execute()
		if err == nil {
			txHash = resp.TxHash
			return
		}
		if resp.Code == 19 {
			return
		}
		if !strings.Contains(err.Error(), "signature verification failed") {
			require.NoError(b.cfg.T, err, "PostOrder: attempt %d", i)
		}
		b.sequence++
		lastErr = err
	}

	b.cfg.T.Fatalf("PostOrder: fail after multiple attempts: %v", lastErr)

	return
}

func executeQuery(req *cliTester.QueryRequest) error {
	const initialTimeoutInMs = 10
	const timeoutMultiplier = 1.05
	const maxRetryAttempts = 50

	var lastErr error
	curRetrySleepDur := time.Duration(initialTimeoutInMs) * time.Millisecond
	for i := 0; i < maxRetryAttempts; i++ {
		output, err := req.Execute()
		if err == nil {
			return nil
		}

		lastErr = err
		if strings.Contains(output, "resource temporarily unavailable") ||
			strings.Contains(output, "connection reset by peer") {
			time.Sleep(curRetrySleepDur)
			curRetrySleepDur = time.Duration(float64(curRetrySleepDur) * timeoutMultiplier)
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

	commonHandler := func(stopFunc func(), ch <-chan coreTypes.ResultEvent, handlerFunc func(coreTypes.ResultEvent)) {
		defer stopFunc()

		for {
			select {
			case <-b.stopCh:
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				handlerFunc(event)
			}
		}
	}

	// post events
	{
		query := fmt.Sprintf("orders.post.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		go commonHandler(stopFunc, ch, b.handleOrderPost)
	}

	// cancel events
	{
		query := fmt.Sprintf("orders.cancel.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		go commonHandler(stopFunc, ch, b.handleOrderCancel)
	}

	// fullyFilled events
	{
		query := fmt.Sprintf("orders.full_fill.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		go commonHandler(stopFunc, ch, b.handleOrderFullFill)
	}

	// partiallyFilled events
	{
		query := fmt.Sprintf("orders.partial_fill.owner='%s'", b.cfg.Address)
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		go commonHandler(stopFunc, ch, b.handleOrderPartialFill)
	}

	// clearance events
	{
		query := fmt.Sprintf("orderbook.clearance.market_id='%s'", b.cfg.MarketID.String())
		stopFunc, ch := b.cfg.Tester.CreateWSConnection(false, b.cfg.Name, query, 1)
		go commonHandler(stopFunc, ch, b.handleOrderBookClearance)
	}
}
