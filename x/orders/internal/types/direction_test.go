// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrders_Direction_Validity(t *testing.T) {
	// ok
	require.True(t, Bid.IsValid())
	require.True(t, Ask.IsValid())

	// fail
	require.False(t, Direction("").IsValid())
	require.False(t, Direction("foo").IsValid())
}
