package binance

import (
	"time"
)

const (
	kSecToMs = 1e3
	kMsToNs  = 1e6
)

// Converts Ticker Unix time in milliseconds to time.Time
// result must be in the close vicinity (+-diffThreshold) of "now" argument
// "now" argument is returned on convert failure
func ConvertTickerUnixMsTime(ms uint64, now time.Time, diffThreshold time.Duration) time.Time {
	if diffThreshold < 0 {
		return now
	}

	sec := int64(ms / kSecToMs)
	ns := int64((ms % kSecToMs) * kMsToNs)
	out := time.Unix(sec, ns)

	minDateTime := now.Add(-diffThreshold)
	maxDateTime := now.Add(diffThreshold)
	if out.Before(minDateTime) || out.After(maxDateTime) {
		return now
	}

	return out
}
