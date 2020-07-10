// +build unit

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOracle_Params(t *testing.T) {
	t.Parallel()

	oracles := []Oracle{NewOracle(sdk.AccAddress([]byte("oracle")))}
	asset := NewAsset("btc_dfi", oracles, true)

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
		params := Params{Assets: []Asset{NewAsset("dfi", oracles, true)}, Nominees: []string{""}}
		require.Error(t, params.Validate())
	}
}
