package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for dumb gas meter.
func TestNewDumbGasMeter(t *testing.T) {
	gasMeter := NewDumbGasMeter()
	require.Zero(t, gasMeter.Limit())
	require.Zero(t, gasMeter.GasConsumed())
	require.False(t, gasMeter.IsPastLimit())
	require.False(t, gasMeter.IsOutOfGas())
	require.Zero(t, gasMeter.GasConsumedToLimit())

	gasMeter.ConsumeGas(100, "test")
	require.Zero(t, gasMeter.GasConsumed())
	require.False(t, gasMeter.IsPastLimit())
	require.False(t, gasMeter.IsOutOfGas())
}
