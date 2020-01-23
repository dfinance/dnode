package keeper

import (
	"bytes"
	"context"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
	"google.golang.org/grpc"
	"net"
	"time"
	"wings-blockchain/cmd/config"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

type Keeper struct {
	cdc      *amino.Codec            // Amino codec.
	storeKey sdk.StoreKey            // Store key.
	client   vm_grpc.VMServiceClient // VM service client.
	listener net.Listener            // VM data server listener.
	config   *config.VMConfig        // VM config.
}

// initialize server
func NewKeeper(storeKey sdk.StoreKey, cdc *amino.Codec, conn *grpc.ClientConn, listener net.Listener, config *config.VMConfig) (keeper Keeper, err error) {
	keeper = Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		client:   vm_grpc.NewVMServiceClient(conn),
		listener: listener,
		config:   config,
	}

	err = StartServer(keeper)
	return
}

func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployContract) (sdk.Events, sdk.Error) {
	timeout := time.Millisecond * time.Duration(keeper.config.DeployTimeout)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewDeployRequest(ctx, msg)
	if sdkErr != nil {
		return nil, sdkErr
	}

	resp, err := keeper.client.ExecuteContracts(connCtx, req)
	if err != nil {
		panic(types.NewErrVMCrashed(err))
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
