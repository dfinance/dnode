// VM gRPC client implementation.
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// VMClient is an aggregated gRPC VM services client.
type VMClient struct {
	vm_grpc.VMCompilerClient
	vm_grpc.VMModulePublisherClient
	vm_grpc.VMScriptExecutorClient
}

// NewVMClient creates VMClient using connection.
func NewVMClient(connection *grpc.ClientConn) VMClient {
	return VMClient{
		VMCompilerClient:        vm_grpc.NewVMCompilerClient(connection),
		VMModulePublisherClient: vm_grpc.NewVMModulePublisherClient(connection),
		VMScriptExecutorClient:  vm_grpc.NewVMScriptExecutorClient(connection),
	}
}

// GetFreeGas returns free gas from execution context.
func GetFreeGas(ctx sdk.Context) sdk.Gas {
	return ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
}

// NewDeployContract creates an object used for publish module requests.
func NewDeployContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte) *vm_grpc.VMPublishModule {
	return &vm_grpc.VMPublishModule{
		Address:      common_vm.Bech32ToLibra(address),
		MaxGasAmount: maxGas,
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
		Address:      common_vm.Bech32ToLibra(address),
		MaxGasAmount: maxGas,
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
		TypeParams:   nil,
		Args:         vmArgs,
	}, nil
}

// NewDeployRequest is a NewDeployContract wrapper: create deploy request.
func NewDeployRequest(ctx sdk.Context, msg types.MsgDeployModule) (*vm_grpc.VMPublishModule, error) {
	return NewDeployContract(msg.Signer, GetFreeGas(ctx), msg.Module), nil
}

// NewExecuteRequest is a NewExecuteContract wrapper: create execute request.
func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteScript, error) {
	contract, err := NewExecuteContract(msg.Signer, GetFreeGas(ctx), msg.Script, msg.Args)
	if err != nil {
		return nil, err
	}

	return contract, nil
}
