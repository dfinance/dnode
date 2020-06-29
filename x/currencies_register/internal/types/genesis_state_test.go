// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test default genesis.
func TestDefaultGenesisState(t *testing.T) {
	dfiVal, _ := sdk.NewIntFromString("100000000000000000000000000")
	defaultGenesis := DefaultGenesisState()
	require.NotNil(t, defaultGenesis)

	require.Len(t, defaultGenesis.Currencies, 1)
	require.Equal(t, "01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d", defaultGenesis.Currencies[0].Path)
	require.Equal(t, "dfi", defaultGenesis.Currencies[0].Denom)
	require.EqualValues(t, 18, defaultGenesis.Currencies[0].Decimals)
	require.EqualValues(t, dfiVal.String(), defaultGenesis.Currencies[0].TotalSupply.String())
}
