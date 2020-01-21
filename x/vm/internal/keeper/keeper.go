package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/go-amino"
	"google.golang.org/grpc"
	"time"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
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

func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployContract) (sdk.Events, sdk.Error) {
	vmAddress := keeper.GetVMAddress(ctx)

	// TODO: secure connection.
	conn, err := grpc.Dial(vmAddress, grpc.WithInsecure())
	if err != nil {
		return nil, types.ErrCantConnectVM(err.Error())
	}

	defer conn.Close()
	client := vm_grpc.NewVMServiceClient(conn)

	timeout := time.Millisecond * keeper.GetVMTimeout(ctx)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewDeployRequest(ctx, msg)
	if err != nil {
		return nil, sdkErr
	}

	resp, err := client.ExecuteContracts(connCtx, req)
	if err != nil {
		return nil, types.ErrDuringVMExec(err.Error())
	}

	events := make(sdk.Events, 0)

	for i, exec := range resp.Executions {
		// TODO: check status and return error in case of errors. Also gas, writeOp, etc.
		for _, value := range exec.WriteSet {
			path := value.GetPath()

			if !bytes.Equal(req.Contracts[i].Address, path.Address) {
				return nil, types.ErrWrongModuleAddress(req.Contracts[i].Address, path.Address)
			}

			if err := keeper.storeModule(ctx, *path, value.Value); err != nil {
				return nil, err
			}

			event := sdk.NewEvent(
				types.EventKeyDeploy,
				sdk.NewAttribute("address", types.DecodeAddress(path.Address).String()),
				sdk.NewAttribute("path", hex.EncodeToString(path.Path)),
			)

			events = append(events, event)
		}
	}

	return events, nil
}
