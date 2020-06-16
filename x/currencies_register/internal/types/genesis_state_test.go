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
	require.Equal(t, "0172c9f1bfe0a2bf6ac342aaa3c3380852d4694ae4e71655d37aa5d2e6700ed94e", defaultGenesis.Currencies[0].Path)
	require.Equal(t, "dfi", defaultGenesis.Currencies[0].Denom)
	require.EqualValues(t, 18, defaultGenesis.Currencies[0].Decimals)
	require.EqualValues(t, dfiVal.String(), defaultGenesis.Currencies[0].TotalSupply.String())
}
