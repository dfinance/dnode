package watcher

import (
	"fmt"
	"math/rand"

	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func (w *Watcher) subscribe() {
	genClientID := func() string {
		return fmt.Sprintf("%d", rand.Uint32())
	}

	commonHandler := func(query string, handlerFunc func(coreTypes.ResultEvent)) {
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, genClientID(), query, 100)
		defer stopFunc()

		for {
			select {
			case <-w.stopCh:
				w.logger.Error(fmt.Sprintf("subscriber crashed: %s", query))
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
	go commonHandler(fmt.Sprintf("orders.post.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue), w.history.HandleOrderPostEvent)

	// cancel events
	go commonHandler(fmt.Sprintf("orders.cancel.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue), w.history.HandleOrderCancelEvent)

	// fullyFilled events
	go commonHandler(fmt.Sprintf("orders.full_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue), w.history.HandleOrderFullFillEvent)

	// partiallyFilled events
	go commonHandler(fmt.Sprintf("orders.partial_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue), w.history.HandleOrderPartialFillEvent)

	// clearance events
	go commonHandler(fmt.Sprintf("orderbook.clearance.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue), w.history.HandleOrderBookClearanceEvent)

	// block events
	go commonHandler("tm.event='NewBlock'", w.history.HandleNewBlockEvent)
}
