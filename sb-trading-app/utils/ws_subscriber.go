package utils

import (
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/libs/log"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

func genSubscriberID() string {
	return fmt.Sprintf("%d", rand.Uint32())
}

func EventsWorker(logger log.Logger, tester *cliTester.CLITester, stopCh chan bool, query string, handlerFunc func(coreTypes.ResultEvent)) {
	var stopFunc func()
	var eventsCh <-chan coreTypes.ResultEvent

	for working := true; working; {
		if eventsCh == nil {
			stopper, ch, err := tester.CreateWSConnection(false, genSubscriberID(), query, 1000)
			if err != nil {
				logger.Error(fmt.Sprintf("subscriber for query %q: connection failed: %v", query, err))
				continue
			}

			logger.Info(fmt.Sprintf("subscriber for query %q: connected", query))
			stopFunc, eventsCh = stopper, ch
		}

		select {
		case <-stopCh:
			working = false
		case event, ok := <-eventsCh:
			if !ok {
				logger.Error(fmt.Sprintf("subscriber for query %q: connection closed, reconnecting", query))
				if stopFunc != nil {
					stopFunc()
				}
				eventsCh = nil
				continue
			}
			handlerFunc(event)
		}
	}

	if stopFunc != nil {
		stopFunc()
	}
	logger.Info(fmt.Sprintf("subscriber for query %q: stopped", query))
}
