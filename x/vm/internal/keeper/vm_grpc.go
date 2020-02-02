package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

func GetFreeGas(ctx sdk.Context) sdk.Gas {
	if ctx.GasMeter().Limit() <= ctx.GasMeter().GasConsumed() {
		return 0
	}

	return ctx.GasMeter().Limit() - ctx.GasMeter().GasConsumed()
}

func NewContract(address sdk.AccAddress, maxGas sdk.Gas, code []byte, contractType vm_grpc.ContractType, args []*vm_grpc.VMArgs) (*vm_grpc.VMContract, sdk.Error) {
	if len(address) != types.VmAddressLength {
		return nil, types.ErrWrongAddressLength(address)
	}

	return &vm_grpc.VMContract{
		Address:      address,
		MaxGasAmount: maxGas,
		GasUnitPrice: types.VmGasPrice,
		Code:         code,
		ContractType: contractType,
		Args:         args,
	}, nil
}

func NewDeployRequest(ctx sdk.Context, msg types.MsgDeployModule) (*vm_grpc.VMExecuteRequest, sdk.Error) {
	address := types.EncodeAddress(msg.Signer)
	gas := GetFreeGas(ctx)

	contract, err := NewContract(address, gas, msg.Module, vm_grpc.ContractType_Module, []*vm_grpc.VMArgs{})
	if err != nil {
		return nil, err
	}

	return &vm_grpc.VMExecuteRequest{
		Contracts: []*vm_grpc.VMContract{contract},
		Options:   0,
	}, nil
}

func NewExecuteRequest(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteRequest, sdk.Error) {
	address := types.EncodeAddress(msg.Signer)
	gas := GetFreeGas(ctx)

	args := make([]*vm_grpc.VMArgs, len(msg.Args))

	for i, arg := range msg.Args {
		args[i] = &vm_grpc.VMArgs{
			Type:  arg.Type,
			Value: arg.Value,
		}
	}

	contract, err := NewContract(address, gas, msg.Script, vm_grpc.ContractType_Script, args)
	if err != nil {
		return nil, err
	}

	return &vm_grpc.VMExecuteRequest{
		Contracts: []*vm_grpc.VMContract{contract},
		Options:   0,
	}, nil
}
