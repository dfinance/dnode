// VM GRPC related functional.
package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Get free gas from execution context.
func GetFreeGas(ctx sdk.Context) sdk.Gas {
	return ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
}

// Create new contract in grpc format for VM request.
func NewContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte, contractType vm_grpc.ContractType, args []*vm_grpc.VMArgs) (*vm_grpc.VMContract, error) {
	return &vm_grpc.VMContract{
		Address:      "0x" + hex.EncodeToString(types.Bech32ToLibra(address)),
		MaxGasAmount: maxGas,
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
		ContractType: contractType,
		Args:         args,
	}, nil
}

// Create deploy request for VM grpc server.
func NewDeployRequest(ctx sdk.Context, msg types.MsgDeployModule) (*vm_grpc.VMExecuteRequest, error) {
	gas := GetFreeGas(ctx)

	contract, err := NewContract(msg.Signer, gas, msg.Module, vm_grpc.ContractType_Module, []*vm_grpc.VMArgs{})
	if err != nil {
		return nil, err
	}

	return &vm_grpc.VMExecuteRequest{
		Contracts: []*vm_grpc.VMContract{contract},
		Options:   0,
	}, nil
}

// Create execute script request for VM grpc server.
func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteRequest, error) {
	gas := GetFreeGas(ctx)

	args := make([]*vm_grpc.VMArgs, len(msg.Args))

	for i, arg := range msg.Args {
		if arg.Type == vm_grpc.VMTypeTag_Address {
			addr, err := sdk.AccAddressFromBech32(arg.Value)
			if err != nil {
				return nil, fmt.Errorf("can't parse address argument %s: %v", arg.Value, err)
			}

			args[i] = &vm_grpc.VMArgs{
				Type:  arg.Type,
				Value: "0x" + hex.EncodeToString(types.Bech32ToLibra(addr)),
			}
		} else {
			args[i] = &vm_grpc.VMArgs{
				Type:  arg.Type,
				Value: arg.Value,
			}
		}
	}

	contract, err := NewContract(msg.Signer, gas, msg.Script, vm_grpc.ContractType_Script, args)
	if err != nil {
		return nil, err
	}

	return &vm_grpc.VMExecuteRequest{
		Contracts: []*vm_grpc.VMContract{contract},
		Options:   0,
	}, nil
}
