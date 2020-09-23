package simulator

import (
	"sort"
	"time"
)

type DefferOps struct {
	DefferQueue []DefferOp
}

type DefferOp struct {
	DFunc     func()
	StartTime time.Time
}

func NewDefferOps() *DefferOps {
	queue := new(DefferOps)
	queue.DefferQueue = make([]DefferOp, 0)
	return queue
}

// Add adds func to deffer queue
func (do *DefferOps) Add(st time.Time, f func()) {
	do.DefferQueue = append(do.DefferQueue, DefferOp{
		DFunc:     f,
		StartTime: st,
	})

	sort.Slice(do.DefferQueue, func(i, j int) bool {
		return do.DefferQueue[i].StartTime.Before(do.DefferQueue[j].StartTime)
	})
}

// Exec executes all operations with time before st
func (do *DefferOps) Exec(st time.Time) {
	for _, op := range do.DefferQueue {
		if op.StartTime.After(st) {
			return
		}
		do.DefferQueue = do.DefferQueue[1:]
		op.DFunc()
	}
}
