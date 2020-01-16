package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/go-amino"
	"google.golang.org/grpc"
	"time"
	vm "wings-blockchain/x/core/protos"
	"wings-blockchain/x/vm/internal/types"
)

var (
	zeroBytes = make([]byte, 12)
)

/*
{"code":[76,73,66,82,65,86,77,10,1,0,8,1,83,0,0,0,4,0,0,0,2,87,0,0,0,4,0,0,0,3,91,0,0,0,3,0,0,0,13,94,0,0,0,10,0,0,0,14,104,0,0,0,5,0,0,0,5,109,0,0,0,24,0,0,0,4,133,0,0,0,64,0,0,0,11,197,0,0,0,10,0,0,0,0,0,1,1,1,2,1,0,0,3,0,2,1,7,0,0,1,7,0,0,0,3,1,7,0,0,8,77,121,77,111,100,117,108,101,9,76,105,98,114,97,67,111,105,110,1,84,2,105,100,213,135,47,194,72,140,54,22,211,54,22,114,175,29,247,218,202,219,159,47,112,121,82,31,112,110,80,47,177,78,179,177,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,2,0,12,0,2]}
*/
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
		Address:      append(sender, zeroBytes...),
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
	// send contract to vm and get output via grpc client
	//	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	vmAddress := keeper.GetVMAddress(ctx)

	// TODO: secure connection.
	conn, err := grpc.Dial(vmAddress, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return types.ErrCantConnectVM(err.Error())
	}

	defer conn.Close()
	client := vm.NewVMServiceClient(conn)
	connCtx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()

	req := NewDeployReq(msg)
	resp, err := client.ExecuteContracts(connCtx, &req)
	if err != nil {
		return types.ErrDuringVMExec(err.Error())
	}

	fmt.Printf("%s", resp.String())

	return nil
}
