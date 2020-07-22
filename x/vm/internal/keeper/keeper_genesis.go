package keeper

import (
	"encoding/hex"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

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

		k.setValue(ctx, accessPath, bzValue)
	}

	// "data" variable can't be used directly as it might contain extra JSON fields
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyGenesis, types.ModuleCdc.MustMarshalJSON(state))
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	k.modulePerms.AutoCheck(types.PermStorageRead)

	store := ctx.KVStore(k.storeKey)
	state := types.GenesisState{}

	if store.Has(types.KeyGenesis) {
		types.ModuleCdc.MustUnmarshalJSON(store.Get(types.KeyGenesis), &state)
	}

	return state
}
