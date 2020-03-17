package binance

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_TickerUnixMsToTime(t *testing.T) {
	now := time.Date(2020, 6, 5, 4, 3, 2, 1000000, time.UTC)

	// check correct input
	{
		input := now
		output := ConvertTickerUnixMsTime(uint64(input.UnixNano()/1e6), now, 1*time.Second)
		require.True(t, output.Equal(input))
	}

	// check correct input (zero threshold, exact match)
	{
		input := now
		output := ConvertTickerUnixMsTime(uint64(input.UnixNano()/1e6), now, 0)
		require.True(t, output.Equal(input))
	}

	// check correct input (within threshold range)
	{
		thresholdLvl := 1 * time.Hour
		input := now.Add(thresholdLvl / 2)
		output := ConvertTickerUnixMsTime(uint64(input.UnixNano()/1e6), now, thresholdLvl)
		require.True(t, output.Equal(input))
	}

	// check incorrect input (min threshold)
	{
		thresholdLvl := 1 * time.Hour
		input := now.Add(-thresholdLvl).Add(-1 * time.Second)
		output := ConvertTickerUnixMsTime(uint64(input.UnixNano()/1e6), now, thresholdLvl)
		require.True(t, output.Equal(now))
	}

	// check incorrect input (max threshold)
	{
		thresholdLvl := 1 * time.Hour
		input := now.Add(thresholdLvl).Add(1 * time.Second)
		output := ConvertTickerUnixMsTime(uint64(input.UnixNano()/1e6), now, thresholdLvl)
		require.True(t, output.Equal(now))
	}
}
