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
	require.Equal(t, "018bfc024222e94fbed60ff0c9c1cf48c5b2809d83c82f513b2c385e21ba8a2d35", defaultGenesis.Currencies[0].Path)
	require.Equal(t, "dfi", defaultGenesis.Currencies[0].Denom)
	require.EqualValues(t, 18, defaultGenesis.Currencies[0].Decimals)
	require.EqualValues(t, dfiVal.String(), defaultGenesis.Currencies[0].TotalSupply.String())
}
