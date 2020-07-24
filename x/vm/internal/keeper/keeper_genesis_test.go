// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

func TestVMKeeper_Genesis(t *testing.T) {
	t.Parallel()

	input := newTestInput(true)
	defer input.Stop()
	ctx, keeper, cdc := input.ctx, input.vk, input.cdc
	store := ctx.KVStore(keeper.storeKey)

	var initState types.GenesisState
	cdc.MustUnmarshalJSON(getGenesis(t), &initState)

	// init
	{
		// genesis flag: before
		require.False(t, store.Has(types.KeyGenesisInit))

		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(initState))

		// genesis flag: after
		require.True(t, store.Has(types.KeyGenesisInit))

		// writeSets
		for _, initWs := range initState.WriteSet {
			initAccessPath, initValue, err := initWs.ToBytes()
			require.NoError(t, err)

			getValue := store.Get(common_vm.GetPathKey(initAccessPath))
			require.NotNil(t, getValue)
			require.EqualValues(t, initValue, getValue)
		}
	}

	// export
	{
		var exportState types.GenesisState
		cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportState)
		// VM genesis also includes CCStorage writeSets
		require.GreaterOrEqual(t, len(exportState.WriteSet), len(initState.WriteSet))

		for _, initWs := range initState.WriteSet {
			foundCnt := 0
			for _, exportWs := range exportState.WriteSet {
				if initWs.Address == exportWs.Address && initWs.Path == exportWs.Path {
					foundCnt++
					require.Equal(t, initWs.Value, exportWs.Value)
				}
			}
			require.Equal(t, 1, foundCnt)
		}
	}
}
