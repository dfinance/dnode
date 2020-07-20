package helpers

import (
	"fmt"
	"runtime"
	"strings"
)

type CallerMeta struct {
	Module string
	Func   string
}

func (c CallerMeta) String() string {
	return fmt.Sprintf("%s::%s", c.Module, c.Func)
}

// Caller returns this func requester and caller meta.
// {shiftFrame} is used when target check func is wrapped into an other func.
// Contract: function shouldn't be moved, as skipFrames must be adjusted
func Caller(shiftFrame uint) (requester, caller CallerMeta) {
	const (
		skipFrames = 1 // skip "runtime.Callers"
	)

	var (
		thisFuncFrameIdx  = int(shiftFrame)       // not used
		requesterFrameIdx = thisFuncFrameIdx + 1  // who called Caller()
		callerFrameIdx    = requesterFrameIdx + 1 // who called callers's Caller()
		endIdx            = callerFrameIdx + 1    // stop parsing here
	)

	// returns moduleName out of the framePath for "x/modules"
	getModuleID := func(value string) string {
		// edge case fix if func called not from "internal" dir
		fixDotNotation := func(moduleName string) string {
			if dotIdx := strings.Index(moduleName, "."); dotIdx != -1 {
				moduleName = moduleName[:dotIdx]
			}
			return moduleName
		}

		// split "github.com/dfinance/x/ccstorage/internal/..."
		items := strings.Split(value, "/")
		for i, item := range items {
			if item == "x" && i+1 < len(items) {
				return fixDotNotation(items[i+1])
			}
		}

		// not a "x/modules" framePath, pick the last module in the path
		if len(items) > 1 {
			return fixDotNotation(items[len(items)-1])
		}

		return "unknownModule"
	}

	// returns functionName out of the framePath
	getFuncID := func(value string) string {
		items := strings.Split(value, ".")
		cnt := len(items)
		if cnt > 0 {
			return items[cnt-1]
		}

		return "unknownFunc"
	}

	// returns metas
	buildID := func(value string) CallerMeta {
		return CallerMeta{
			Module: getModuleID(value),
			Func:   getFuncID(value),
		}
	}

	requesterValue, callerValue := "", ""
	defer func() {
		requester = buildID(requesterValue)
		caller = buildID(callerValue)
	}()

	programCounters := make([]uintptr, endIdx)
	framesCnt := runtime.Callers(skipFrames, programCounters)
	if framesCnt < callerFrameIdx+1 {
		return
	}

	// cycle through all frames and pick target ones by index
	frames := runtime.CallersFrames(programCounters[:framesCnt])
	for frameIdx, hasNextFrame := 0, true; hasNextFrame; frameIdx++ {
		var curFrame runtime.Frame
		curFrame, hasNextFrame = frames.Next()

		switch frameIdx {
		case requesterFrameIdx:
			requesterValue = curFrame.Function
		case callerFrameIdx:
			callerValue = curFrame.Function
		case endIdx:
			hasNextFrame = false
		}
	}

	return
}
