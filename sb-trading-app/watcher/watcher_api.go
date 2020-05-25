package watcher

import (
	"fmt"

	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func (w *Watcher) subscribe() {
	commonHandler := func(stopFunc func(), ch <-chan coreTypes.ResultEvent, handlerFunc func(coreTypes.ResultEvent)) {
		defer stopFunc()

		for {
			select {
			case <-w.stopCh:
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
		query := fmt.Sprintf("orders.post.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 100)
		go commonHandler(stopFunc, ch, w.history.HandleOrderPostEvent)
	}

	// cancel events
	{
		query := fmt.Sprintf("orders.cancel.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		go commonHandler(stopFunc, ch, w.history.HandleOrderCancelEvent)
	}

	// fullyFilled events
	{
		query := fmt.Sprintf("orders.full_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		go commonHandler(stopFunc, ch, w.history.HandleOrderFullFillEvent)
	}

	// partiallyFilled events
	{
		query := fmt.Sprintf("orders.partial_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		go commonHandler(stopFunc, ch, w.history.HandleOrderPartialFillEvent)
	}

	// clearance events
	{
		query := fmt.Sprintf("orderbook.clearance.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		go commonHandler(stopFunc, ch, w.history.HandleOrderBookClearanceEvent)
	}
}
