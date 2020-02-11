// VM keeper processing messages from handler.
package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"net"
	"time"
	"wings-blockchain/cmd/config"
	"wings-blockchain/x/core"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
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

// Initialize VM keeper (include grpc client to VM and grpc server for data store).
func NewKeeper(storeKey sdk.StoreKey, cdc *amino.Codec, conn *grpc.ClientConn, listener net.Listener, config *config.VMConfig, logger log.Logger) (keeper Keeper) {
	keeper = Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		rawClient: conn,
		client:    vm_grpc.NewVMServiceClient(conn),
		listener:  listener,
		config:    config,
	}

	keeper.dsServer = NewDSServer(&keeper, logger.With("module", "ds-server"))
	keeper.rawDSServer = StartServer(keeper.listener, keeper.dsServer)

	return
}

// Stop DS server and close connection to VM.
func (keeper Keeper) CloseConnections() {
	if keeper.rawDSServer != nil {
		keeper.rawDSServer.Stop()
	}

	if keeper.rawClient != nil {
		keeper.rawClient.Close()
	}
}

// Execute script.
func (keeper Keeper) ExecuteScript(ctx sdk.Context, msg types.MsgExecuteScript) (sdk.Events, sdk.Error) {
	timeout := time.Millisecond * time.Duration(keeper.config.TimeoutExecute)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return nil, sdkErr
	}

	dumbGasCtx := ctx.WithGasMeter(core.NewDumbGasMeter())
	keeper.dsServer.SetContext(&dumbGasCtx)

	resp, err := keeper.client.ExecuteContracts(connCtx, req)
	if err != nil {
		fmt.Printf("error is: %s\n", err.Error())
		panic(types.NewErrVMCrashed(err))
	}

	keeper.dsServer.SetContext(nil)

	if len(resp.Executions) != 1 {
		// error because execution amount during such transaction could be only one.
		return nil, types.ErrWrongExecutionResponse(*resp)
	}

	exec := resp.Executions[0]
	events := keeper.processExecution(ctx, exec)

	return events, nil
}

// Deploy module.
func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployModule) (sdk.Events, sdk.Error) {
	timeout := time.Millisecond * time.Duration(keeper.config.TimeoutDeploy)
	connCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, sdkErr := NewDeployRequest(ctx, msg)
	if sdkErr != nil {
		return nil, sdkErr
	}

	dumbGasCtx := ctx.WithGasMeter(core.NewDumbGasMeter())
	keeper.dsServer.SetContext(&dumbGasCtx)

	resp, err := keeper.client.ExecuteContracts(connCtx, req)
	if err != nil {
		panic(types.NewErrVMCrashed(err))
	}

	keeper.dsServer.SetContext(nil)

	if len(resp.Executions) != 1 {
		// error because execution amount during such transaction could be only one.
		return nil, types.ErrWrongExecutionResponse(*resp)
	}

	exec := resp.Executions[0]
	events := keeper.processExecution(ctx, exec)

	return events, nil
}
