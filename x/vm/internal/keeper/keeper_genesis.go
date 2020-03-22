// Work with genesis data.
package keeper

import (
	"encoding/hex"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Process genesis state and write state.
func (keeper Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	var state types.GenesisState

	types.ModuleCdc.MustUnmarshalJSON(data, &state)

	for _, genWriteOp := range state.WriteSet {
		bzAddr, err := hex.DecodeString(genWriteOp.Address)
		if err != nil {
			helpers.CrashWithError(err)
		}

		bzPath, err := hex.DecodeString(genWriteOp.Path)
		if err != nil {
			helpers.CrashWithError(err)
		}

		bzValue, err := hex.DecodeString(genWriteOp.Value)
		if err != nil {
			helpers.CrashWithError(err)
		}

		accessPath := &vm_grpc.VMAccessPath{
			Address: bzAddr,
			Path:    bzPath,
		}

		keeper.setValue(ctx, accessPath, bzValue)
	}

	// "data" variable can't be used directly as it might contain extra JSON fields
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.KeyGenesis, types.ModuleCdc.MustMarshalJSON(state))
}

func (keeper Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	store := ctx.KVStore(keeper.storeKey)
	state := types.GenesisState{}

	if store.Has(types.KeyGenesis) {
		types.ModuleCdc.MustUnmarshalJSON(store.Get(types.KeyGenesis), &state)
	}

	return state
}
