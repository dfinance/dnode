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

	stateWriteSetPaths := make([]vm_grpc.VMAccessPath, 0, len(state.WriteSet))
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
		stateWriteSetPaths = append(stateWriteSetPaths, *accessPath)
	}

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.KeyGenesisInitialized, []byte{1})
	store.Set(types.KeyGenesisWriteSetPaths, types.ModuleCdc.MustMarshalJSON(stateWriteSetPaths))
}

func (keeper Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	state := types.GenesisState{}
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.KeyGenesisInitialized) {
		return state
	}
	if !store.Has(types.KeyGenesisWriteSetPaths) {
		return state
	}

	var writeSetPaths []vm_grpc.VMAccessPath
	types.ModuleCdc.MustUnmarshalJSON(store.Get(types.KeyGenesisWriteSetPaths), &writeSetPaths)
	for _, path := range writeSetPaths {
		writeSet := types.GenesisWriteOp{
			Address: hex.EncodeToString(path.Address),
			Path:    hex.EncodeToString(path.Path),
			Value:   hex.EncodeToString(keeper.getValue(ctx, &path)),
		}
		state.WriteSet = append(state.WriteSet, writeSet)
	}

	return state
}
