package watcher

import (
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func (w *Watcher) subscribe() func(){
	// post events
	var postStopFunc func()
	{
		query := fmt.Sprintf("orders.post.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 100)
		postStopFunc = stopFunc

		go func() {
			for {
				if event, ok := <-ch; ok {
					w.history.HandleOrderPostEvent(event)
				} else {
					return
				}
			}
		}()
	}

	// cancel events
	var cancelStopFunc func()
	{
		query := fmt.Sprintf("orders.cancel.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		cancelStopFunc = stopFunc

		go func() {
			for {
				if event, ok := <-ch; ok {
					w.history.HandleOrderCancelEvent(event)
				} else {
					return
				}
			}
		}()
	}

	// fullyFilled events
	var fullFillStopFunc func()
	{
		query := fmt.Sprintf("orders.full_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		fullFillStopFunc = stopFunc

		go func() {
			for {
				if event, ok := <-ch; ok {
					w.history.HandleOrderFullFillEvent(event)
				} else {
					return
				}
			}
		}()
	}

	// partiallyFilled events
	var partialFillStopFunc func()
	{
		query := fmt.Sprintf("orders.partial_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		partialFillStopFunc = stopFunc

		go func() {
			for {
				if event, ok := <-ch; ok {
					w.history.HandleOrderPartialFillEvent(event)
				} else {
					return
				}
			}
		}()
	}

	// clearance events
	var clearanceStopFunc func()
	{
		query := fmt.Sprintf("orderbook.clearance.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue)
		stopFunc, ch := w.cfg.Tester.CreateWSConnection(false, "watcher", query, 1)
		clearanceStopFunc = stopFunc

		go func() {
			for {
				if event, ok := <-ch; ok {
					w.history.HandleOrderBookClearanceEvent(event)
				} else {
					return
				}
			}
		}()
	}

	return func() {
		postStopFunc()
		cancelStopFunc()
		fullFillStopFunc()
		partialFillStopFunc()
		clearanceStopFunc()
	}
}
