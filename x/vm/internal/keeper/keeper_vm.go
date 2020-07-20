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
	"github.com/tendermint/tendermint/crypto/tmhash"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Set value in storage by access path.
func (keeper Keeper) setValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(keeper.storeKey)
	key := common_vm.MakePathKey(accessPath)

	store.Set(key, value)
}

// Check if VM storage has value by access path.
func (keeper Keeper) HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	keeper.modulePerms.AutoCheck(types.PermStorageReader)

	return keeper.hasValue(ctx, accessPath)
}

// Public get value by path.
func (keeper Keeper) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	keeper.modulePerms.AutoCheck(types.PermStorageReader)

	return keeper.getValue(ctx, accessPath)
}

// GetValue with middleware context-dependant processing.
func (keeper Keeper) GetValueWithMiddlewares(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	keeper.modulePerms.AutoCheck(types.PermStorageReader)

	for _, f := range keeper.dsServer.dataMiddlewares {
		data, err := f(ctx, accessPath)
		if err != nil {
			return nil
		}
		if data != nil {
			return data
		}
	}

	return keeper.GetValue(ctx, accessPath)
}

// Public set value.
func (keeper Keeper) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	keeper.modulePerms.AutoCheck(types.PermStorageWriter)

	keeper.setValue(ctx, accessPath, value)
}

// Delete value.
func (keeper Keeper) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	keeper.modulePerms.AutoCheck(types.PermStorageWriter)

	keeper.delValue(ctx, accessPath)
}

// Public get path for oracle price.
func (keeper Keeper) GetOracleAccessPath(assetCode dnTypes.AssetCode) *vm_grpc.VMAccessPath {
	keeper.modulePerms.AutoCheck(types.PermStorageReader)

	seed := xxhash.NewS64(0)
	if _, err := seed.WriteString(strings.ToLower(assetCode.String())); err != nil {
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
	if _, err := hash.Write(bz); err != nil {
		panic(err)
	}
	path := hash.Sum(tag)

	return &vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    path,
	}
}

// Check if vm storage contains key.
func (keeper Keeper) hasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(keeper.storeKey)
	key := common_vm.MakePathKey(accessPath)

	return store.Has(key)
}

// Get value from storage by access path.
func (keeper Keeper) getValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(keeper.storeKey)
	key := common_vm.MakePathKey(accessPath)

	return store.Get(key)
}

// Delete key in storage by access path.
func (keeper Keeper) delValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(keeper.storeKey)
	key := common_vm.MakePathKey(accessPath)

	store.Delete(key)
}

// Process result of VM module/script execution.
func (keeper Keeper) processExecution(ctx sdk.Context, exec *vm_grpc.VMExecuteResponse) {
	// consume gas, if execution took too much gas - panic and mark transaction as out of gas
	ctx.GasMeter().ConsumeGas(exec.GasUsed, "vm script/module execution")

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(types.ModuleName))
	ctx.EventManager().EmitEvents(types.NewContractEvents(exec))

	// process "keep" status
	if exec.Status == vm_grpc.ContractStatus_Keep {
		// return on "error" status
		if exec.StatusStruct != nil && exec.StatusStruct.MajorStatus != types.VMCodeExecuted {
			types.PrintVMStackTrace(tmhash.Sum(ctx.TxBytes()), keeper.Logger(ctx), exec)
			return
		}

		keeper.processWriteSet(ctx, exec.WriteSet)

		// emit VM events (panic on "out of gas", emitted events stays in the EventManager)
		for _, vmEvent := range exec.Events {
			ctx.EventManager().EmitEvent(types.NewMoveEvent(ctx.GasMeter(), vmEvent))
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
			panic(fmt.Errorf("unknown write op, couldn't happen: %d", value.Type))
		}
	}
}
