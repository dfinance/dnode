// VM gRPC client implementation.
package keeper

import (
	"google.golang.org/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/compiler_grpc"
	"github.com/dfinance/dvm-proto/go/metadata_grpc"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	VMMaxGasLimit = ^uint64(0)/1000 - 1
)

// VMClient is an aggregated gRPC VM services client.
type VMClient struct {
	VMCompilerClient        compiler_grpc.DvmCompilerClient
	VMMetaDataClient        metadata_grpc.DVMBytecodeMetadataClient
	VMModulePublisherClient vm_grpc.VMModulePublisherClient
	VMScriptExecutorClient  vm_grpc.VMScriptExecutorClient
}

// NewVMClient creates VMClient using connection.
func NewVMClient(connection *grpc.ClientConn) VMClient {
	return VMClient{
		VMCompilerClient:        compiler_grpc.NewDvmCompilerClient(connection),
		VMMetaDataClient:        metadata_grpc.NewDVMBytecodeMetadataClient(connection),
		VMModulePublisherClient: vm_grpc.NewVMModulePublisherClient(connection),
		VMScriptExecutorClient:  vm_grpc.NewVMScriptExecutorClient(connection),
	}
}

// GetFreeGas returns free gas from execution context.
func GetFreeGas(ctx sdk.Context) sdk.Gas {
	return ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
}

// getVMLimitedGas returns gas less than VM max limit.
func getVMLimitedGas(gas sdk.Gas) sdk.Gas {
	if gas > VMMaxGasLimit {
		return VMMaxGasLimit
	}
	return gas
}

// NewDeployContract creates an object used for publish module requests.
func NewDeployContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte) *vm_grpc.VMPublishModule {
	return &vm_grpc.VMPublishModule{
		Sender:       common_vm.Bech32ToLibra(address),
		MaxGasAmount: getVMLimitedGas(maxGas),
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
	}
}

// NewExecuteContract creates an object used for script execute requests.
func NewExecuteContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte, args []types.ScriptArg) (*vm_grpc.VMExecuteScript, error) {
	vmArgs := make([]*vm_grpc.VMArgs, len(args))
	for argIdx, arg := range args {
		vmArgs[argIdx] = &vm_grpc.VMArgs{
			Type:  arg.Type,
			Value: arg.Value,
		}
	}

	return &vm_grpc.VMExecuteScript{
		Senders:      [][]byte{common_vm.Bech32ToLibra(address)},
		MaxGasAmount: getVMLimitedGas(maxGas),
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
		TypeParams:   nil,
		Args:         vmArgs,
	}, nil
}

// NewDeployRequest is a NewDeployContract wrapper: create deploy request.
func NewDeployRequest(ctx sdk.Context, signer sdk.AccAddress, contract types.Contract) *vm_grpc.VMPublishModule {
	return NewDeployContract(signer, GetFreeGas(ctx), contract)
}

// NewExecuteRequest is a NewExecuteContract wrapper: create execute request.
func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteScript, error) {
	contract, err := NewExecuteContract(msg.Signer, GetFreeGas(ctx), msg.Script, msg.Args)
	if err != nil {
		return nil, err
	}

	return contract, nil
}
