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
		curTimeout := time.Duration(req.CurrentTimeout)*time.Millisecond
		connCtx, _ := context.WithTimeout(context.Background(), curTimeout)

		resp, err := keeper.client.ExecuteContracts(connCtx, req.Raw)
		if err != nil {
			if req.Attempt == 0 {
				// write to Sentry (if enabled)
				keeper.Logger(ctx).Error(fmt.Sprintf("Can't get answer from VM in %v, will try to reconnect in %s attempts: %v", req.CurrentTimeout, GetMaxAttemptsStr(req.MaxAttempts), err))
			}
			req.Attempt += 1

			if req.MaxAttempts != 0 && req.Attempt == req.MaxAttempts {
				// return error because of max attempts.
				logErr := fmt.Errorf("max %d attemps reached, can't get answer from VM: %v", req.Attempt, err)
				keeper.Logger(ctx).Error(logErr.Error())
				return nil, logErr
			}
			time.Sleep(curTimeout)

			req.CurrentTimeout += int(math.Round(float64(req.CurrentTimeout) * keeper.config.BackoffMultiplier))
			if req.CurrentTimeout > keeper.config.MaxBackoff {
				req.CurrentTimeout = keeper.config.MaxBackoff
			}

			continue
		}
		keeper.Logger(ctx).Info(fmt.Sprintf("Successfully connected to VM with %v timeout in %d attempts", req.CurrentTimeout, req.Attempt))

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
	}

	return keeper.retryExecReq(ctx, retryReq)
}

// Convert max attempts amount to string representation.
func GetMaxAttemptsStr(maxAttempts int) string {
	if maxAttempts == 0 {
		return "infinity"
	} else {
		return fmt.Sprintf("%d", maxAttempts)
	}
}
