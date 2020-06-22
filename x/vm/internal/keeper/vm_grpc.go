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

// Create deploy request for VM gRPC server.
func NewDeployRequest(ctx sdk.Context, msg types.MsgDeployModule) (*vm_grpc.VMPublishModule, error) {
	return &vm_grpc.VMPublishModule{
		Address:      common_vm.Bech32ToLibra(msg.Signer),
		MaxGasAmount: GetFreeGas(ctx),
		GasUnitPrice: types.VmGasPrice,
		Code:         msg.Module,
	}, nil
}

// Create execute script request for VM gRPC server.
func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteScript, error) {
	vmArgs := make([]*vm_grpc.VMArgs, len(msg.Args))

	for argIdx, arg := range msg.Args {
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
		Address:      common_vm.Bech32ToLibra(msg.Signer),
		MaxGasAmount: GetFreeGas(ctx),
		GasUnitPrice: types.VmGasPrice,
		Code:         msg.Script,
		TypeParams:   nil,
		Args:         nil,
	}, nil
}
