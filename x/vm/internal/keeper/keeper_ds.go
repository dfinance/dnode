package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// RetryExecReq contains VM "execution" request meta (request details and retry settings).
type RetryExecReq struct {
	// Request to retry (module publish).
	RawModule *vm_grpc.VMPublishModule
	// Request to retry (script execution)
	RawScript *vm_grpc.VMExecuteScript
	// Max number of request attempts (0 - infinite)
	MaxAttempts uint
	// Request timeout per attempt (0 - infinite) [ms]
	ReqTimeoutInMs uint
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
	curAttempt := uint(0)
	reqTimeout := time.Duration(req.ReqTimeoutInMs) * time.Millisecond
	reqStartedAt := time.Now()

	go func() {
		var connCancel context.CancelFunc

		cancelPrevConn := func() {
			if connCancel != nil {
				connCancel()
				connCancel = nil
			}
		}

		defer func() {
			cancelPrevConn()
			close(doneCh)
		}()

		for {
			var resp *vm_grpc.VMExecuteResponse
			var err error

			curAttempt++
			cancelPrevConn()

			connCtx := context.Background()
			if reqTimeout > 0 {
				connCtx, connCancel = context.WithTimeout(context.Background(), reqTimeout)
			}

			if req.RawModule != nil {
				resp, err = k.client.PublishModule(connCtx, req.RawModule)
			} else if req.RawScript != nil {
				resp, err = k.client.ExecuteScript(connCtx, req.RawScript)
			}

			if err == nil {
				retResp, retErr = resp, nil
				return
			}

			if curAttempt == req.MaxAttempts {
				retResp, retErr = nil, err
				return
			}
		}
	}()
	<-doneCh

	reqDur := time.Now().Sub(reqStartedAt)
	msg := fmt.Sprintf("in %d attempt(s) with %v timeout (%v)", curAttempt, reqTimeout, reqDur)
	if retErr == nil {
		k.GetLogger(ctx).Info(fmt.Sprintf("Successfull VM request (%s)", msg))
	} else {
		k.GetLogger(ctx).Error(fmt.Sprintf("Failed VM request (%s): %v", msg, retErr))
		retErr = fmt.Errorf("%s: %w", msg, retErr)
	}

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
		MaxAttempts:    k.config.MaxAttempts,
		ReqTimeoutInMs: k.config.ReqTimeoutInMs,
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
