// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Plan(t *testing.T) {
	require.NoError(t, NewPlan(1).ValidateBasic())
	require.Error(t, NewPlan(-1).ValidateBasic())
	require.Error(t, NewPlan(0).ValidateBasic())
}
