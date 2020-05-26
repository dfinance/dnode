package bot

import (
	"fmt"
	"strings"
	"sync"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

type ApiCli struct {
	sync.Mutex
	tester         *cliTester.CLITester
	accountAddress string
	accountNumber  uint64
	sequenceNumber uint64
	marketID       dnTypes.ID
	baseDenom      string
	quoteDenom     string
	orderTtlInSec  int
}

func (a *ApiCli) GetAccount() (sequence uint64, baseBalance, quoteBalance sdk.Uint, retErr error) {
	baseBalance, quoteBalance = sdk.ZeroUint(), sdk.ZeroUint()

	q, acc := a.tester.QueryAccount(a.accountAddress)
	if err := a.executeQuery(q); err != nil {
		retErr = fmt.Errorf("QueryAccount: %w", err)
		return
	}

	sequence = acc.Sequence
	for _, coin := range acc.Coins {
		if coin.Denom == a.baseDenom {
			baseBalance = sdk.Uint(coin.Amount)
			continue
		}
		if coin.Denom == a.quoteDenom {
			quoteBalance = sdk.Uint(coin.Amount)
			continue
		}
	}

	return
}

func (a *ApiCli) PostOrder(price, quantity sdk.Uint, direction orderTypes.Direction) (txHash string, retErr error) {
	const maxRetries = 10

	if price.IsZero() || quantity.IsZero() {
		return
	}

	a.Lock()
	defer a.Unlock()

	defer func() {
		if txHash != "" {
			a.sequenceNumber++
		}
	}()

	sendTx := func() (verificationFailed, stop bool, unhandledErr error) {
		txCmd := a.tester.TxOrdersPost(a.accountAddress, a.marketID, direction, price, quantity, a.orderTtlInSec)
		txCmd.DisableBroadcastMode()
		txCmd.SetAccountNumber(a.accountNumber)
		txCmd.SetSequenceNumber(a.sequenceNumber)

		resp, err := txCmd.Execute()
		if err == nil {
			txHash = resp.TxHash
			return
		}

		if resp.Code == 19 {
			stop = true
			return
		}
		if strings.Contains(err.Error(), "signature verification failed") {
			verificationFailed = true
			return
		}
		unhandledErr = err

		return
	}

	for i := 0; i < maxRetries; i++ {
		doSeqDrop, doStop, err := sendTx()

		if err != nil {
			retErr = fmt.Errorf("PostOrder: attempt %d: %w", i, err)
			return
		}
		if doStop {
			return
		}
		if doSeqDrop {
			if i == 0 {
				if sequence, _, _, err := a.GetAccount(); err != nil {
					retErr = fmt.Errorf("PostOrder: attempt %d: GetAccount: %w", i, err)
					return
				} else {
					a.sequenceNumber = sequence
				}
			} else {
				a.sequenceNumber++
			}
			continue
		}

		return
	}
	retErr = fmt.Errorf("PostOrder: sequenceNumber fit failed")

	return
}

func (a *ApiCli) GetOrder(id dnTypes.ID) (order *orderTypes.Order, retErr error) {
	q, orderPtr := a.tester.QueryOrdersOrder(id)
	if err := a.executeQuery(q); err != nil {
		if strings.Contains(err.Error(), "wrong orderID") {
			return
		}

		retErr = fmt.Errorf("GetOrder: %w", err)
		return
	}
	order = orderPtr

	return
}

func (a *ApiCli) executeQuery(req *cliTester.QueryRequest) error {
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

func (a *ApiCli) executeTx(req *cliTester.TxRequest) (string, error) {
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

func NewApiCli(tester *cliTester.CLITester, accNumber uint64, accAddress string, marketID dnTypes.ID, baseDenom, quoteDenom string, orderTtlInSec int) *ApiCli {
	return &ApiCli{
		tester:         tester,
		accountAddress: accAddress,
		accountNumber:  accNumber,
		marketID:       marketID,
		baseDenom:      baseDenom,
		quoteDenom:     quoteDenom,
		orderTtlInSec:  orderTtlInSec,
	}
}
