// Work with genesis data.
package keeper

import (
	"encoding/hex"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
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

	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.KeyGenesisInitialized, []byte{1})
}
