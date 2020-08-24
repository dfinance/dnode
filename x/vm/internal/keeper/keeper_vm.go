package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// HasValue checks if VM storage has writeSet data by accessPath.
func (k Keeper) HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	k.modulePerms.AutoCheck(types.PermStorageRead)

	return k.hasValue(ctx, accessPath)
}

// GetValue returns VM storage writeSet data by accessPath.
func (k Keeper) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	k.modulePerms.AutoCheck(types.PermStorageRead)

	return k.getValue(ctx, accessPath)
}

// GetValueWithMiddlewares extends GetValue with middleware context-dependant processing.
func (k Keeper) GetValueWithMiddlewares(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	k.modulePerms.AutoCheck(types.PermStorageRead)

	for _, f := range k.dsServer.dataMiddlewares {
		data, err := f(ctx, accessPath)
		if err != nil {
			return nil
		}
		if data != nil {
			return data
		}
	}

	return k.GetValue(ctx, accessPath)
}

// SetValue sets VM storage writeSet data by accessPath.
func (k Keeper) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	k.modulePerms.AutoCheck(types.PermStorageWrite)

	k.setValue(ctx, accessPath, value)
}

// DelValue removes VM storage writeSet data by accessPath.
func (k Keeper) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	k.modulePerms.AutoCheck(types.PermStorageWrite)

	k.delValue(ctx, accessPath)
}

// hasValue checks that VM storage contains key.
func (k Keeper) hasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(k.storeKey)
	key := common_vm.GetPathKey(accessPath)

	return store.Has(key)
}

// getValue returns value from VM storage by key.
func (k Keeper) getValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(k.storeKey)
	key := common_vm.GetPathKey(accessPath)

	return store.Get(key)
}

// setValue sets value to VM storage by key.
func (k Keeper) setValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(k.storeKey)
	key := common_vm.GetPathKey(accessPath)

	store.Set(key, value)
}

// delValue removes value from VM storage by key.
func (k Keeper) delValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(k.storeKey)
	key := common_vm.GetPathKey(accessPath)

	store.Delete(key)
}

// processExecution processes VM execution result (emit events, convert VM events, update writeSets).
func (k Keeper) processExecution(ctx sdk.Context, exec *vm_grpc.VMExecuteResponse) {
	// consume gas, if execution took too much gas - panic and mark transaction as out of gas
	ctx.GasMeter().ConsumeGas(exec.GasUsed, "vm script/module execution")

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(types.ModuleName))
	ctx.EventManager().EmitEvents(types.NewContractEvents(exec))

	// process success status
	if exec.GetStatus().GetError() == nil {
		k.processWriteSet(ctx, exec.WriteSet)

		// emit VM events (panic on "out of gas", emitted events stays in the EventManager)
		for _, vmEvent := range exec.Events {
			ctx.EventManager().EmitEvent(types.NewMoveEvent(ctx.GasMeter(), vmEvent))
		}
	}
}

// processWriteSet processes VM execution writeSets (set/delete).
func (k Keeper) processWriteSet(ctx sdk.Context, writeSet []*vm_grpc.VMValue) {
	for _, value := range writeSet {
		// check type and solve what to do.
		if value.Type == vm_grpc.VmWriteOp_Deletion {
			// deleting key.
			k.delValue(ctx, value.Path)
		} else if value.Type == vm_grpc.VmWriteOp_Value {
			// write to storage.
			k.setValue(ctx, value.Path, value.Value)
		} else {
			// must not happens at all
			panic(fmt.Errorf("unknown write op, couldn't happen: %d", value.Type))
		}
	}
}

// iterateOverValues iterates over all VM values and processes them with handler (stop when handler returns false).
func (k Keeper) iterateOverValues(ctx sdk.Context, handler func(accessPath *vm_grpc.VMAccessPath, value []byte) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, common_vm.GetPathPrefixKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		accessPath := common_vm.MustParsePathKey(iterator.Key())
		value := iterator.Value()

		if !handler(accessPath, value) {
			break
		}
	}
}
