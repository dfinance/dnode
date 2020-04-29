// VM and storage related things.
package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/OneOfOne/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/crypto/sha3"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Set value in storage by access path.
func (keeper Keeper) setValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(accessPath)

	store.Set(key, value)
}

// Public get value by path.
func (keeper Keeper) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	return keeper.getValue(ctx, accessPath)
}

// Public set value.
func (keeper Keeper) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	keeper.setValue(ctx, accessPath, value)
}

// Delete value.
func (keeper Keeper) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	keeper.delValue(ctx, accessPath)
}

// Public get path for oracle price.
func (keeper Keeper) GetOracleAccessPath(assetCode string) *vm_grpc.VMAccessPath {
	seed := xxhash.NewS64(0)
	_, err := seed.WriteString(strings.ToLower(assetCode))
	if err != nil {
		panic(err)
	}

	ticketPair := seed.Sum64()

	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, ticketPair)
	tag, err := hex.DecodeString("ff")
	if err != nil {
		panic(err)
	}

	hash := sha3.New256()
	hash.Write(bz)
	path := hash.Sum(tag)

	return &vm_grpc.VMAccessPath{
		Address: make([]byte, types.VmAddressLength),
		Path:    path,
	}
}

// Get value from storage by access path.
func (keeper Keeper) getValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(accessPath)

	return store.Get(key)
}

// Check if storage has value by access path.
func (keeper Keeper) hasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(accessPath)

	return store.Has(key)
}

// Delete key in storage by access path.
func (keeper Keeper) delValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(keeper.storeKey)
	key := types.MakePathKey(accessPath)

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
