package keeper

import (
	"bytes"
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/go-amino"
	"google.golang.org/grpc"
	"time"
	vm "wings-blockchain/x/core/protos"
	"wings-blockchain/x/vm/internal/types"
)

type Keeper struct {
	cdc        *amino.Codec
	storeKey   sdk.StoreKey
	paramStore params.Subspace
}

func NewKeeper(storeKey sdk.StoreKey, cdc *amino.Codec, paramStore params.Subspace) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramStore: paramStore.WithKeyTable(NewKeyTable()),
	}
}

func NewContract(sender sdk.AccAddress, maxGasAmount uint64, code types.Contract, contractType vm.ContractType, args []*vm.VMArgs) vm.VMContract {
	return vm.VMContract{
		Address:      types.EncodeAddress(sender),
		MaxGasAmount: maxGasAmount,
		GasUnitPrice: 1,
		Code:         code,
		ContractType: contractType,
		Args:         args,
	}
}

func NewDeployReq(msg types.MsgDeployContract) vm.VMExecuteRequest {
	contract := NewContract(msg.Signer, 100000000, msg.Contract, 0, []*vm.VMArgs{})

	return vm.VMExecuteRequest{
		Contracts: []*vm.VMContract{&contract},
		Options:   0,
	}
}

func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployContract) sdk.Error {
	vmAddress := keeper.GetVMAddress(ctx)

	// TODO: secure connection.
	conn, err := grpc.Dial(vmAddress, grpc.WithInsecure())
	if err != nil {
		return types.ErrCantConnectVM(err.Error())
	}

	defer conn.Close()
	client := vm.NewVMServiceClient(conn)

	timeout := time.Millisecond * keeper.GetVMTimeout(ctx)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req := NewDeployReq(msg)
	resp, err := client.ExecuteContracts(connCtx, &req)
	if err != nil {
		return types.ErrDuringVMExec(err.Error())
	}

	for i, exec := range resp.Executions {
		// TODO: check status and return error in case of errors. Also gas, writeOp, etc.
		for _, value := range exec.WriteSet {
			path := value.GetPath()

			if !bytes.Equal(req.Contracts[i].Address, path.Address) {
				return types.ErrWrongModuleAddress(req.Contracts[i].Address, path.Address)
			}

			if err := keeper.storeModule(ctx, *path, value.Value); err != nil {
				return err
			}
		}
	}

	return nil
}
