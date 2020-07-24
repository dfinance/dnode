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

// RetryExecReq contains VM "execution" request meta (request details and retry settings).
type RetryExecReq struct {
	RawModule      *vm_grpc.VMPublishModule // Request to retry (module publish).
	RawScript      *vm_grpc.VMExecuteScript // Request to retry (script execution)
	Attempt        int                      // Current attempt.
	CurrentTimeout int                      // Current timeout.
	MaxAttempts    int                      // Max attempts.
}

// StartDSServer starts DataSource server.
func (k *Keeper) StartDSServer(ctx sdk.Context) {
	k.modulePerms.AutoCheck(types.PermDsAdmin)

	// check if genesis initialized
	// if no - skip, it will be started later.
	store := ctx.KVStore(k.storeKey)
	if store.Has(types.KeyGenesisInit) && !k.dsServer.IsStarted() {
		// launch server.
		k.rawDSServer = StartServer(k.listener, k.dsServer)
	}
}

// SetDSContext sets DataSource server context (storage context should be updated periodically to provide actual data).
func (k Keeper) SetDSContext(ctx sdk.Context) {
	k.modulePerms.AutoCheck(types.PermDsAdmin)

	k.dsServer.SetContext(ctx.WithGasMeter(types.NewDumbGasMeter()))
}

// CloseConnections stops DataSource server and close connection to VM.
func (k Keeper) CloseConnections() {
	k.modulePerms.AutoCheck(types.PermDsAdmin)

	if k.rawDSServer != nil {
		k.rawDSServer.Stop()
	}

	if k.rawClient != nil {
		k.rawClient.Close()
	}
}

// retryExecReq sends request with retry mechanism and waits for connection and execution.
// Contract: either RawModule or RawScript must be specified for RetryExecReq.
func (k Keeper) retryExecReq(ctx sdk.Context, req RetryExecReq) (retResp *vm_grpc.VMExecuteResponse, retErr error) {
	doneCh := make(chan bool)
	go func() {
		var cancelCtx func()
		stopPrevCtx := func() {
			if cancelCtx != nil {
				cancelCtx()
			}
		}

		defer func() {
			close(doneCh)
			stopPrevCtx()
		}()

		for {
			stopPrevCtx()
			curTimeout := time.Duration(req.CurrentTimeout) * time.Millisecond
			connCtx, cancelFunc := context.WithTimeout(context.Background(), curTimeout)
			cancelCtx = cancelFunc

			connStartedAt := time.Now()
			var resp *vm_grpc.VMExecuteResponse
			var err error
			if req.RawModule != nil {
				resp, err = k.client.PublishModule(connCtx, req.RawModule)
			} else if req.RawScript != nil {
				resp, err = k.client.ExecuteScript(connCtx, req.RawScript)
			}

			connDuration := time.Since(connStartedAt)
			if err != nil {
				if req.Attempt == 0 {
					// write to Sentry (if enabled)
					k.GetLogger(ctx).Error(fmt.Sprintf("Can't get answer from VM in %v, will try to reconnect in %s attempts: %v", req.CurrentTimeout, getMaxAttemptsStr(req.MaxAttempts), err))
				}
				req.Attempt += 1

				if req.MaxAttempts != 0 && req.Attempt == req.MaxAttempts {
					// return error because of max attempts.
					logErr := fmt.Errorf("max %d attemps reached, can't get answer from VM: %v", req.Attempt, err)
					k.GetLogger(ctx).Error(logErr.Error())
					retErr = logErr
					return
				}

				if curTimeout > connDuration {
					time.Sleep(curTimeout - connDuration)
				}

				req.CurrentTimeout += int(math.Round(float64(req.CurrentTimeout) * k.config.BackoffMultiplier))
				if req.CurrentTimeout > k.config.MaxBackoff {
					req.CurrentTimeout = k.config.MaxBackoff
				}

				continue
			}
			k.GetLogger(ctx).Info(fmt.Sprintf("Successfully connected to VM with %v timeout in %d attempts", req.CurrentTimeout, req.Attempt))
			retResp = resp

			return
		}
	}()
	<-doneCh

	return
}

// sendExecuteReq sends request with retry mechanism.
func (k Keeper) sendExecuteReq(ctx sdk.Context, moduleReq *vm_grpc.VMPublishModule, scriptReq *vm_grpc.VMExecuteScript) (*vm_grpc.VMExecuteResponse, error) {
	if moduleReq == nil && scriptReq == nil {
		return nil, fmt.Errorf("request (module / script) not specified")
	}
	if moduleReq != nil && scriptReq != nil {
		return nil, fmt.Errorf(" only single request (module / script) is supported")
	}

	retryReq := RetryExecReq{
		RawModule:      moduleReq,
		RawScript:      scriptReq,
		CurrentTimeout: k.config.InitialBackoff,
		MaxAttempts:    k.config.MaxAttempts,
	}

	if k.config.MaxAttempts < 0 {
		// just send, in case of error - return error and panic.
		retryReq.MaxAttempts = 1
	}

	return k.retryExecReq(ctx, retryReq)
}

// getMaxAttemptsStr converts max attempts amount to string representation.
func getMaxAttemptsStr(maxAttempts int) string {
	if maxAttempts == 0 {
		return "infinity"
	} else {
		return fmt.Sprintf("%d", maxAttempts)
	}
}
