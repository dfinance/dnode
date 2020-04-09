// +build unit

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloatToFPString(t *testing.T) {
	a := 0.1
	res := FloatToFPString(a, Precision)
	require.Equal(t, res, "10000000")

	a = 1.0
	res = FloatToFPString(a, Precision)
	require.Equal(t, res, "100000000")

	a = 0.0
	res = FloatToFPString(a, Precision)
	require.Equal(t, res, "0")

	a = 0.00000001
	res = FloatToFPString(a, Precision)
	require.Equal(t, res, "1")

	a = 0.000000001
	res = FloatToFPString(a, Precision)
	require.Equal(t, res, "0")

	a = 0
	res = FloatToFPString(a, Precision)
	require.Equal(t, res, "0")
}
