package watcher

import (
	"fmt"
	"time"
)

func (w *Watcher) Work() {
	w.subscribe()

	workDur := time.Duration(w.cfg.WorkDurtInSec) * time.Second
	w.logger.Info(fmt.Sprintf("starting for %v", workDur))
	for _, m := range w.marketStates {
		for _, b := range m.bots {
			w.wg.Add(1)
			go b.Start(w.wg, w.stopCh)
		}
	}

	stopCh := time.After(workDur)
	tickCh := time.Tick(10 * time.Second)
	for working := true; working; {
		select {
		case <-tickCh:
			//for _, market := range w.marketStates {
			//	balances := make(HistoryBalances, 0, len(market.bots))
			//	for _, bot := range market.bots {
			//		baseBalance, quoteBalance := bot.Balances()
			//		balances = append(balances, HistoryBalance{
			//			Name:  bot.Name(),
			//			Base:  baseBalance,
			//			Quote: quoteBalance,
			//		})
			//	}
			//
			//	w.history.SetCurBalances(market.id.String(), balances)
			//}

			w.history.ResetCurItem()
		case <-stopCh:
			w.logger.Info("stopping")
			working = false

			close(w.stopCh)
			w.wg.Wait()

			w.logger.Info(fmt.Sprintf("results:\n%s", w.history.String(true, true)))
		}
	}
}
