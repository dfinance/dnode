// VM GRPC related functional.
package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"

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
		var argValue []byte

		switch arg.Type {
		case vm_grpc.VMTypeTag_Address:
			addr, err := sdk.AccAddressFromBech32(arg.Value)
			if err != nil {
				return nil, fmt.Errorf("argument[%d]: can't parse address argument %s: %v", argIdx, arg.Value, err)
			}
			argValue = common_vm.Bech32ToLibra(addr)
		case vm_grpc.VMTypeTag_U8:
			value, err := strconv.ParseUint(arg.Value, 10, 8)
			if err != nil {
				return nil, fmt.Errorf("argument[%d]: can't parse u8 argument %s: %v", argIdx, arg.Value, err)
			}
			argValue = []byte{uint8(value)}
		case vm_grpc.VMTypeTag_U64:
			value, err := strconv.ParseUint(arg.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("argument[%d]: can't parse u64 argument %s: %v", argIdx, arg.Value, err)
			}
			argValue = make([]byte, 8)
			binary.LittleEndian.PutUint64(argValue, value)
		case vm_grpc.VMTypeTag_U128:
			value := sdk.NewUintFromString(arg.Value)
			if value.BigInt().BitLen() > 128 {
				return nil, fmt.Errorf("argument[%d]: can't parse u128 argument %s: invalid bitLen %d", argIdx, arg.Value, value.BigInt().BitLen())
			}

			// BigInt().Bytes() returns BigEndian format, reverse it
			argValue = value.BigInt().Bytes()
			for left, right := 0, len(argValue)-1; left < right; left, right = left+1, right-1 {
				argValue[left], argValue[right] = argValue[right], argValue[left]
			}
			// Extend to 16 bytes
			if len(argValue) < 16 {
				zeros := make([]byte, 16-len(argValue))
				argValue = append(argValue, zeros...)
			}
		case vm_grpc.VMTypeTag_Bool:
			value, err := strconv.ParseBool(arg.Value)
			if err != nil {
				return nil, fmt.Errorf("argument[%d]: can't parse bool argument %s: %v", argIdx, arg.Value, err)
			}
			if value {
				argValue = []byte{1}
			} else {
				argValue = []byte{0}
			}
		default:
			argValue = []byte(arg.Value)
		}

		vmArgs[argIdx] = &vm_grpc.VMArgs{
			Type:  arg.Type,
			Value: argValue,
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
