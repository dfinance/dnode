// Keeper methods related to data source.
package keeper

import (
	"context"
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Start Data source (DS) server.
func (keeper *Keeper) StartDSServer(ctx sdk.Context) {
	// check if genesis initialized
	// if no - skip, it will be started later.
	store := ctx.KVStore(keeper.storeKey)
	if store.Has(types.KeyGenesis) && !keeper.dsServer.IsStarted() {
		// launch server.
		keeper.rawDSServer = StartServer(keeper.listener, keeper.dsServer)
	}
}

// Set DS (data-source) server context.
func (keeper Keeper) SetDSContext(ctx sdk.Context) {
	keeper.dsServer.SetContext(ctx.WithGasMeter(types.NewDumbGasMeter()))
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

type RetryExecReq struct {
	Raw            *vm_grpc.VMExecuteRequest // Request to retry.
	Attempt        int                       // Current attempt.
	CurrentTimeout int                       // Current timeout.
	MaxAttempts    int                       // Max attempts.
}

func (keeper Keeper) retryExecReq(ctx sdk.Context, req *RetryExecReq) (*vm_grpc.VMExecuteResponses, error) {
	connCtx, cancel := context.WithTimeout(context.Background(), time.Duration(req.CurrentTimeout)*time.Millisecond)
	defer cancel()

	resp, err := keeper.client.ExecuteContracts(connCtx, req.Raw)

	if err != nil {
		// write to sentry
		keeper.Logger(ctx).Error(fmt.Sprintf("can't get answer from vm in %d ms, retry #%d"))

		if req.MaxAttempts != 0 {
			req.Attempt += 1

			if req.Attempt == req.MaxAttempts {
				// return error because of max attempts.
				return nil, fmt.Errorf("max attempts reached: %d", req.Attempt)
			}
		}

		req.CurrentTimeout += int(math.Round(float64(req.CurrentTimeout) * keeper.config.BackoffMultiplier))

		if req.CurrentTimeout > keeper.config.MaxBackoff {
			req.CurrentTimeout = keeper.config.MaxBackoff
		}

		return keeper.retryExecReq(ctx, req)
	}

	return resp, nil
}

// Send request with retry mechanism.
func (keeper Keeper) sendExecuteReq(ctx sdk.Context, req *vm_grpc.VMExecuteRequest) (*vm_grpc.VMExecuteResponses, error) {
	retryReq := RetryExecReq{
		Raw:            req,
		CurrentTimeout: keeper.config.InitialBackoff,
		MaxAttempts:    keeper.config.MaxAttempts,
	}

	if keeper.config.MaxAttempts < 0 {
		// just send, in case of error - return error and panic.
		return keeper.retryExecReq(ctx, &retryReq)
	} else {
		retryReq.MaxAttempts = 1
		return keeper.retryExecReq(ctx, &retryReq)
	}
}
