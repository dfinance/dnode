// Work with genesis data.
package keeper

import (
	"encoding/hex"
	"encoding/json"

	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Process genesis state and write state.
func (keeper Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	var state types.GenesisState

	types.ModuleCdc.MustUnmarshalJSON(data, &state)

	for _, genWriteOp := range state.WriteSet {
		bzAddr, err := hex.DecodeString(genWriteOp.Address)
		if err != nil {
			panic(err)
		}

		bzPath, err := hex.DecodeString(genWriteOp.Path)
		if err != nil {
			panic(err)
		}

		bzValue, err := hex.DecodeString(genWriteOp.Value)
		if err != nil {
			panic(err)
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
