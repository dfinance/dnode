package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	var state types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &state)

	for genWOIdx, genWriteOp := range state.WriteSet {
		accessPath, value, err := genWriteOp.ToBytes()
		if err != nil {
			panic(fmt.Errorf("writeSetOp[%d]: %w", genWOIdx, err))
		}

		k.setValue(ctx, accessPath, value)
	}

	// raise flag for DS server that genesis was inited
	store := ctx.KVStore(k.storeKey)
	store.Set(types.KeyGenesisInit, []byte{0x1})
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermStorageRead)

	state := types.GenesisState{}
	k.iterateOverValues(ctx, func(accessPath *vm_grpc.VMAccessPath, value []byte) bool {
		writeSetOp := types.GenesisWriteOp{
			Address: hex.EncodeToString(accessPath.Address),
			Path:    hex.EncodeToString(accessPath.Path),
			Value:   hex.EncodeToString(value),
		}
		state.WriteSet = append(state.WriteSet, writeSetOp)

		return true
	})

	return k.cdc.MustMarshalJSON(state)
}
