package watcher

import (
	"fmt"
	"time"
)

func (w *Watcher) Work() {
	w.subscribe()

	workDur := time.Duration(w.cfg.WorkDurtInSec) * time.Second
	w.logger.Info(fmt.Sprintf("starting for %v with %d clients", workDur, w.curBots))
	for _, m := range w.marketStates {
		for i := uint(0); i < w.curBots; i++ {
			b := m.bots[i]
			w.wg.Add(1)
			go b.Start(w.wg, w.stopCh)
		}
	}

	stopCh := time.After(workDur)
	botAddTicker := time.NewTicker(workDur / time.Duration(w.cfg.MaxBots-w.cfg.MinBots+1))
	historyTicker := time.NewTicker(10 * time.Second)
	for working := true; working; {
		select {
		case <-historyTicker.C:
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
		case <-botAddTicker.C:
			if botsDiff := w.cfg.MaxBots - w.curBots; botsDiff > 0 {
				w.logger.Info(fmt.Sprintf("adding client: %d left", botsDiff-1))
				for _, m := range w.marketStates {
					w.wg.Add(1)
					go m.bots[w.curBots].Start(w.wg, w.stopCh)
				}
				w.curBots++
				w.history.SetCurBots(w.curBots)
			}
		case <-stopCh:
			w.logger.Info("stopping")
			working = false

			close(w.stopCh)
			w.wg.Wait()
			time.Sleep(500 * time.Millisecond)

			w.logger.Info(fmt.Sprintf("results:\n%s", w.history.String(true, false, true)))
		}
	}
}
