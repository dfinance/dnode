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

type ApiRest struct {
	sync.Mutex
	tester         *cliTester.CLITester
	accountName    string
	accountAddress sdk.AccAddress
	accountNumber  uint64
	sequenceNumber uint64
	marketID       dnTypes.ID
	baseDenom      string
	quoteDenom     string
	orderTtlInSec  int
}

func (a *ApiRest) GetAccount() (sequence uint64, baseBalance, quoteBalance sdk.Uint, retErr error) {
	baseBalance, quoteBalance = sdk.ZeroUint(), sdk.ZeroUint()

	req, acc := a.tester.RestQueryAuthAccount(a.accountAddress.String())
	if err := a.executeQuery(req); err != nil {
		retErr = fmt.Errorf("GetAccount: %w", err)
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

func (a *ApiRest) PostOrder(price, quantity sdk.Uint, direction orderTypes.Direction) (txHash string, retErr error) {
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
		assetCode := dnTypes.AssetCode(a.baseDenom + "_" + a.quoteDenom)
		req, tx := a.tester.RestTxOrdersPostOrderRaw(
			a.accountName,
			a.accountAddress,
			a.accountNumber,
			a.sequenceNumber,
			assetCode,
			direction,
			price,
			quantity,
			uint64(a.orderTtlInSec),
		)

		err := req.Execute()
		if err == nil {
			txHash = tx.TxHash
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

func (a *ApiRest) GetOrder(id dnTypes.ID) (order *orderTypes.Order, retErr error) {
	req, orderPtr := a.tester.RestQueryOrder(id)
	if err := a.executeQuery(req); err != nil {
		if strings.Contains(err.Error(), "wrong orderID") {
			return
		}

		retErr = fmt.Errorf("GetOrder: %w", err)
		return
	}
	order = orderPtr

	return
}

func (a *ApiRest) executeQuery(req *cliTester.RestRequest) error {
	const initialTimeoutInMs = 10
	const timeoutMultiplier = 1.05
	const maxRetryAttempts = 50

	var lastErr error
	curRetrySleepDur := time.Duration(initialTimeoutInMs) * time.Millisecond
	for i := 0; i < maxRetryAttempts; i++ {
		err := req.Execute()
		if err == nil {
			return nil
		}

		lastErr = err
		if strings.Contains(err.Error(), "cannot assign requested address") {
			time.Sleep(curRetrySleepDur)
			curRetrySleepDur = time.Duration(float64(curRetrySleepDur) * timeoutMultiplier)
			continue
		}

		return err
	}

	return fmt.Errorf("retry failed: %v", lastErr)
}

func NewApiRest(tester *cliTester.CLITester, accNumber uint64, accName, accAddress string, marketID dnTypes.ID, baseDenom, quoteDenom string, orderTtlInSec int) *ApiRest {
	addr, _ := sdk.AccAddressFromBech32(accAddress)

	return &ApiRest{
		tester:         tester,
		accountName:    accName,
		accountAddress: addr,
		accountNumber:  accNumber,
		marketID:       marketID,
		baseDenom:      baseDenom,
		quoteDenom:     quoteDenom,
		orderTtlInSec:  orderTtlInSec,
	}
}
