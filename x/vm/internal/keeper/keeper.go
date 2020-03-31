// VM keeper processing messages from handler.
package keeper

import (
	"context"
	"fmt"
	"net"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// VM keeper.
type Keeper struct {
	cdc      *amino.Codec // Amino codec.
	storeKey sdk.StoreKey // Store key.

	client    vm_grpc.VMServiceClient // VM service client.
	listener  net.Listener            // VM data server listener.
	rawClient *grpc.ClientConn        // GRPC connection to VM.

	config *config.VMConfig // VM config.

	dsServer    *DSServer    // Data-source server.
	rawDSServer *grpc.Server // GRPC raw server.
}

// Check that VMStorage is compatible with keeper (later we can do it by events probably).
var _ common_vm.VMStorage = Keeper{}

// Initialize VM keeper (include grpc client to VM and grpc server for data store).
func NewKeeper(storeKey sdk.StoreKey, cdc *amino.Codec, conn *grpc.ClientConn, listener net.Listener, config *config.VMConfig) (keeper Keeper) {
	keeper = Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		rawClient: conn,
		client:    vm_grpc.NewVMServiceClient(conn),
		listener:  listener,
		config:    config,
	}

	keeper.dsServer = NewDSServer(&keeper)
	return
}

// VM keeper logger.
func (Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "vm")
}

// Execute script.
func (keeper Keeper) ExecuteScript(ctx sdk.Context, msg types.MsgExecuteScript) sdk.Error {
	timeout := time.Millisecond * time.Duration(keeper.config.TimeoutExecute)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	resp, err := keeper.client.ExecuteContracts(connCtx, req)
	if err != nil {
		keeper.Logger(ctx).Error(fmt.Sprintf("grpc error: %s", err.Error()))
		panic(types.NewErrVMCrashed(err))
	}

	if len(resp.Executions) != 1 {
		// error because execution amount during such transaction could be only one.
		return types.ErrWrongExecutionResponse(*resp)
	}

	exec := resp.Executions[0]
	keeper.processExecution(ctx, exec)

	return nil
}

// Deploy module.
func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployModule) sdk.Error {
	timeout := time.Millisecond * time.Duration(keeper.config.TimeoutDeploy)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewDeployRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	resp, err := keeper.client.ExecuteContracts(connCtx, req)
	if err != nil {
		keeper.Logger(ctx).Error(fmt.Sprintf("grpc error: %s", err.Error()))
		panic(types.NewErrVMCrashed(err))
	}

	if len(resp.Executions) != 1 {
		// error because execution amount during such transaction could be only one.
		return types.ErrWrongExecutionResponse(*resp)
	}

	exec := resp.Executions[0]
	keeper.processExecution(ctx, exec)

	return nil
}
