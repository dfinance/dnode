// VM and storage related things.
package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// Set value in storage by access path.
func (keeper Keeper) setValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(*accessPath)

	store.Set(key, value)
}

// Get value from storage by access path.
func (keeper Keeper) getValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(*accessPath)

	return store.Get(key)
}

// Check if storage has value by access path.
func (keeper Keeper) hasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(*accessPath)

	return store.Has(key)
}

// Delete key in storage by access path.
func (keeper Keeper) delValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(*accessPath)

	store.Delete(key)
}

// Process result of VM module/script execution.
func (keeper Keeper) processExecution(ctx sdk.Context, exec *vm_grpc.VMExecuteResponse) {
	// consume gas, if execution took too much gas - panic and mark transaction as out of gas.
	ctx.GasMeter().ConsumeGas(exec.GasUsed, "vm script/module execution")

	// process status
	if exec.Status == vm_grpc.ContractStatus_Discard {
		ctx.EventManager().EmitEvent(types.NewEventDiscard(exec.StatusStruct))
	} else {
		ctx.EventManager().EmitEvent(types.NewEventKeep())

		if exec.StatusStruct != nil && exec.StatusStruct.MajorStatus != types.VMCodeExecuted {
			ctx.EventManager().EmitEvent(types.NewEventError(exec.StatusStruct))
		}

		// processing write set.
		keeper.processWriteSet(ctx, exec.WriteSet)

		for _, vmEvent := range exec.Events {
			ctx.EventManager().EmitEvent(types.NewEventFromVM(vmEvent))
		}
	}
}

// Process write set of module/script execution.
func (keeper Keeper) processWriteSet(ctx sdk.Context, writeSet []*vm_grpc.VMValue) {
	for _, value := range writeSet {
		// check type and solve what to do.
		if value.Type == vm_grpc.VmWriteOp_Deletion {
			// deleting key.
			keeper.delValue(ctx, value.Path)
		} else if value.Type == vm_grpc.VmWriteOp_Value {
			// write to storage.
			keeper.setValue(ctx, value.Path, value.Value)
		} else {
			// must not happens at all
			panic(fmt.Sprintf("Unknown write op, couldn't happen: %d", value.Type))
		}
	}
}
