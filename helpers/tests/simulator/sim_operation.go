package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SimOperationNextExecFn returns next execution time for SimOperation.
type SimOperationNextExecFn func(curTime time.Time, period time.Duration) time.Time

// NewPeriodicNextExecFn is a periodic execution.
func NewPeriodicNextExecFn() SimOperationNextExecFn {
	return func(curTime time.Time, period time.Duration) time.Time {
		return curTime.Add(period)
	}
}

// SimOperationHandler handles operation using Simulator infra.
type SimOperationHandler func(s *Simulator) bool

// SimOperation keeps operation state and handlers.
// CONTRACT: operation must update changed Simulator state (account balance, modified validator, new delegation, etc).
type SimOperation struct {
	handlerFn    SimOperationHandler
	nextExecFn   SimOperationNextExecFn
	period       time.Duration
	nextExecTime time.Time
	execCounter  int
}

// Exec executes the operation if its time has come.
func (op *SimOperation) Exec(s *Simulator, curTime time.Time) (executed bool) {
	defer func() {
		if !executed {
			return
		}

		op.nextExecTime = op.nextExecFn(curTime, op.period)
		op.execCounter++
	}()

	if op.nextExecTime.IsZero() {
		executed = true
		op.execCounter--
		return
	}

	if curTime.After(op.nextExecTime) {
		executed = op.handlerFn(s)
	}

	return
}

// NewSimOperation creates a new SimOperation.
func NewSimOperation(period time.Duration, nextExecFn SimOperationNextExecFn, handlerFn SimOperationHandler) *SimOperation {
	return &SimOperation{
		handlerFn:  handlerFn,
		nextExecFn: nextExecFn,
		period:     period,
	}
}

// checkRatioArg checks SimOperation ratio coef input (0 < value <= 1.0).
func checkRatioArg(opName, argName string, argValue sdk.Dec) {
	errMsgPrefix := fmt.Sprintf("%s: %s: ", opName, argName)
	if argValue.LTE(sdk.ZeroDec()) {
		panic(fmt.Errorf("%s: LTE 0", errMsgPrefix))
	}
	if argValue.GT(sdk.OneDec()) {
		panic(fmt.Errorf("%s: GE 1", errMsgPrefix))
	}
}
