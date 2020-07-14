package watcher

import (
	"fmt"
	"strings"

	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/dfinance/dnode/helpers/tests/sb-trading-app/utils"
	"github.com/dfinance/dnode/x/orderbook"
	"github.com/dfinance/dnode/x/orders"
)

func (w *Watcher) subscribe() {
	ordersHandler := func(event coreTypes.ResultEvent) {
		for attrKey := range event.Events {
			if strings.HasPrefix(attrKey, orders.EventTypeOrderPost) {
				w.history.HandleOrderPostEvent(event)
				continue
			}

			if strings.HasPrefix(attrKey, orders.EventTypeOrderCancel) {
				w.history.HandleOrderCancelEvent(event)
				continue
			}

			if strings.HasPrefix(attrKey, orders.EventTypeFullyFilledOrder) {
				w.history.HandleOrderFullFillEvent(event)
				continue
			}

			if strings.HasPrefix(attrKey, orders.EventTypePartiallyFilledOrder) {
				w.history.HandleOrderPartialFillEvent(event)
				continue
			}
		}
	}

	// subscribe to all orders events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("message.module='%s'", orders.ModuleName),
		ordersHandler,
	)

	orderbookHandler := func(event coreTypes.ResultEvent) {
		for attrKey := range event.Events {
			if strings.HasPrefix(attrKey, orderbook.EventTypeClearance) {
				w.history.HandleOrderBookClearanceEvent(event)
				continue
			}
		}
	}

	// subscribe to all orderbook events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("message.module='%s'", orderbook.ModuleName),
		orderbookHandler,
	)

	// block events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		"tm.event='NewBlock'",
		w.history.HandleNewBlockEvent,
	)
}
