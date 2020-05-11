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

// Retry "execution" request.
// Contains information about request to VM and retry settings.
type RetryExecReq struct {
	Raw            *vm_grpc.VMExecuteRequest // Request to retry.
	Attempt        int                       // Current attempt.
	CurrentTimeout int                       // Current timeout.
	MaxAttempts    int                       // Max attempts.
}

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

// Send request with retry mechanism and wait for connection and execution or return error.
func (keeper Keeper) retryExecReq(ctx sdk.Context, req RetryExecReq) (*vm_grpc.VMExecuteResponses, error) {
	for {
		connCtx, cancel := context.WithTimeout(context.Background(), time.Duration(req.CurrentTimeout)*time.Millisecond)
		defer cancel()

		resp, err := keeper.client.ExecuteContracts(connCtx, req.Raw)

		if err != nil {
			// Write to sentry.
			if req.Attempt == 0 {
				keeper.Logger(ctx).Error(fmt.Sprintf("can't get answer from vm in %d ms, will try to reconnect in %s attempts", req.CurrentTimeout, GetMaxAttemptsStr(req.MaxAttempts)))
			}

			req.Attempt += 1

			if req.MaxAttempts != 0 && req.Attempt == req.MaxAttempts {
				// return error because of max attempts.
				err := fmt.Errorf("max %d attemps reached, can't get answer from VM", req.Attempt)
				keeper.Logger(ctx).Error(err.Error())
				return nil, err
			}

			req.CurrentTimeout += int(math.Round(float64(req.CurrentTimeout) * keeper.config.BackoffMultiplier))

			if req.CurrentTimeout > keeper.config.MaxBackoff {
				req.CurrentTimeout = keeper.config.MaxBackoff
			}

			time.Sleep(1 * time.Millisecond)
			continue
		}

		keeper.Logger(ctx).Info(fmt.Sprintf("successful connected to vm with %d ms timeout", req.CurrentTimeout))
		return resp, nil
	}
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
		retryReq.MaxAttempts = 1
		return keeper.retryExecReq(ctx, retryReq)
	} else {
		return keeper.retryExecReq(ctx, retryReq)
	}
}

// Convert max attempts amount to string representation.
func GetMaxAttemptsStr(maxAttempts int) string {
	if maxAttempts == 0 {
		return "infinity"
	} else {
		return fmt.Sprintf("%d", maxAttempts)
	}
}
