// VM GRPC related functional.
package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// VMClient is an aggregated gRPC services client.
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

// Get free gas from execution context.
func GetFreeGas(ctx sdk.Context) sdk.Gas {
	return ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
}

// NewDeployContract creates object used for publish module request.
func NewDeployContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte) *vm_grpc.VMPublishModule {
	return &vm_grpc.VMPublishModule{
		Address:      common_vm.Bech32ToLibra(address),
		MaxGasAmount: maxGas,
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
	}
}

// NewExecuteContract creates object used for script execute request.
func NewExecuteContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte, args []types.ScriptArg) (*vm_grpc.VMExecuteScript, error) {
	vmArgs := make([]*vm_grpc.VMArgs, len(args))
	for argIdx, arg := range args {
		if arg.Type == vm_grpc.VMTypeTag_Address {
			addr, err := sdk.AccAddressFromBech32(arg.Value)
			if err != nil {
				return nil, fmt.Errorf("argument[%d]: can't parse address argument %s: %v", argIdx, arg.Value, err)
			}

			vmArgs[argIdx] = &vm_grpc.VMArgs{
				Type:  arg.Type,
				Value: common_vm.Bech32ToLibra(addr),
			}
		} else {
			vmArgs[argIdx] = &vm_grpc.VMArgs{
				Type:  arg.Type,
				Value: []byte(arg.Value),
			}
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

// Create deploy request for VM gRPC server.
func NewDeployRequest(ctx sdk.Context, msg types.MsgDeployModule) (*vm_grpc.VMPublishModule, error) {
	return NewDeployContract(msg.Signer, GetFreeGas(ctx), msg.Module), nil
}

// Create execute script request for VM gRPC server.
func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteScript, error) {
	contract, err := NewExecuteContract(msg.Signer, GetFreeGas(ctx), msg.Script, msg.Args)
	if err != nil {
		return nil, err
	}

	return contract, nil
}
