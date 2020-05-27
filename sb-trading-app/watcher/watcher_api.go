package watcher

import (
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/sb-trading-app/utils"
)

func (w *Watcher) subscribe() {
	// post events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("orders.post.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		w.history.HandleOrderPostEvent,
	)

	// cancel events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("orders.cancel.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		w.history.HandleOrderCancelEvent,
	)

	// fullyFilled events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("orders.full_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		w.history.HandleOrderFullFillEvent,
	)

	// partiallyFilled events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("orders.partial_fill.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		w.history.HandleOrderPartialFillEvent,
	)

	// clearance events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		fmt.Sprintf("orderbook.clearance.%s='%s'", dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		w.history.HandleOrderBookClearanceEvent,
	)

	// block events
	go utils.EventsWorker(w.logger, w.cfg.Tester, w.stopCh,
		"tm.event='NewBlock'",
		w.history.HandleNewBlockEvent,
	)
}
