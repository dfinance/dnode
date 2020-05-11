package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// Test init genesis.
func TestKeeper_InitGenesis(t *testing.T) {
	input := GetTestInput(t)

	defaultGenesis := types.DefaultGenesisState()
	require.NotNil(t, defaultGenesis)

	bz, err := input.keeper.cdc.MarshalJSON(defaultGenesis)
	require.NoError(t, err)

	err = input.keeper.InitGenesis(input.ctx, bz)
	require.NoError(t, err)

	info, err := input.keeper.GetCurrencyInfo(input.ctx, defaultGenesis.Currencies[0].Denom)
	require.NoError(t, err)

	dfiCurr := defaultGenesis.Currencies[0]

	require.EqualValues(t, dfiCurr.Denom, info.Denom)
	require.EqualValues(t, dfiCurr.Decimals, info.Decimals)
	require.EqualValues(t, dfiCurr.TotalSupply.String(), info.TotalSupply.String())
	require.False(t, info.IsToken)
}

// Test export genesis.
func TestKeeper_ExportGenesis(t *testing.T) {
	input := GetTestInput(t)

	defaultGenesis := types.DefaultGenesisState()
	require.NotNil(t, defaultGenesis)

	bz, err := input.keeper.cdc.MarshalJSON(defaultGenesis)
	require.NoError(t, err)

	err = input.keeper.InitGenesis(input.ctx, bz)
	require.NoError(t, err)

	exported := input.keeper.ExportGenesis(input.ctx)
	require.EqualValues(t, bz, exported)
}
