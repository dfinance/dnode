// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Check Params validate method.
func TestOracle_Params_Valid(t *testing.T) {
	t.Parallel()

	oracles := []Oracle{NewOracle(sdk.AccAddress([]byte("oracle")))}
	asset := NewAsset("btc_xfi", oracles, true)

	// ok
	{
		params := Params{Assets: []Asset{asset}, Nominees: []string{"nominee"}}
		require.NoError(t, params.Validate())
	}

	// fail nominee
	{
		params := Params{Assets: []Asset{asset}, Nominees: []string{""}}
		require.Error(t, params.Validate())
	}

	// fail asset
	{
		params := Params{Assets: []Asset{NewAsset("xfi", oracles, true)}, Nominees: []string{""}}
		require.Error(t, params.Validate())
	}
}
