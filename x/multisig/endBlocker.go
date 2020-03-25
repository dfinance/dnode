// End blocker implementation.
package multisig

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/poa"
)

// Implements end blocker to process active calls and their confirmations.
func EndBlocker(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper) []abci.Event {
	logger := keeper.getLogger(ctx)
	resEvents := sdk.NewEventManager()

	start := ctx.BlockHeight() - keeper.GetIntervalToExecute(ctx)

	if start < 0 {
		start = 0
	}

	// Iterate active calls
	activeIterator := keeper.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
	defer activeIterator.Close()
	resEvents.EmitEvent(sdk.NewEvent("start-active-calls-ex", sdk.Attribute{Key: "height", Value: fmt.Sprintf("%d", start)}))
	for ; activeIterator.Valid(); activeIterator.Next() {
		bs := activeIterator.Value()

		var callId uint64
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bs, &callId)

		confirmations, err := keeper.GetConfirmations(ctx, callId)
		if err != nil {
			if err.Codespace() == types.DefaultCodespace && err.Code() == types.CodeErrWrongCallId {
				continue
			}

			panic(err)
		}

		// check if call is confirmed enough
		if uint16(confirmations) >= poaKeeper.GetEnoughConfirmations(ctx) {
			resEvents.EmitEvent(sdk.NewEvent("execute-call", sdk.Attribute{Key: "callId", Value: fmt.Sprintf("%d", callId)}))
			// call confirmed - execute
			call := keeper.getCallById(ctx, callId)
			call.Approved = true

			handler := keeper.router.GetRoute(call.Msg.Route())

			cacheCtx, writeCache := ctx.CacheContext()
			err := handler(cacheCtx, call.Msg)

			if err != nil {
				// call execution failed, write it to status
				call.Failed = true
				call.Error = err.Error()

				resEvents.EmitEvent(sdk.NewEvent("failed", sdk.Attribute{Key: "callId", Value: fmt.Sprintf("%d", callId)}))

				logger.Info(
					fmt.Sprintf("Failed execution of %d call, error: %s, marked as failed",
						callId, err.Error()),
				)
			} else {
				call.Executed = true
				writeCache()

				resEvents.EmitEvent(sdk.NewEvent("executed", sdk.Attribute{Key: "callId", Value: fmt.Sprintf("%d", callId)}))

				logger.Info(
					fmt.Sprintf("Call %d executed completed", callId),
				)
			}

			// save call as executed
			keeper.saveCallById(ctx, callId, call)

			// remove proposal from queue
			keeper.removeCallFromQueue(ctx, callId, call.Height)
		}
	}

	if start > keeper.GetIntervalToExecute(ctx) {
		resEvents.EmitEvent(sdk.NewEvent("start-rejected-calls-rem", sdk.Attribute{
			Key:   "callId",
			Value: fmt.Sprintf("%d", start),
		}))

		// Remove not confirmed calls during intervals
		rejectedIterator := keeper.GetQueueIteratorTill(ctx, start)
		defer rejectedIterator.Close()
		for ; rejectedIterator.Valid(); rejectedIterator.Next() {
			bs := rejectedIterator.Value()

			var callId uint64
			keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bs, &callId)

			call := keeper.getCallById(ctx, callId)
			call.Rejected = true

			keeper.saveCallById(ctx, callId, call)
			keeper.removeCallFromQueue(ctx, callId, call.Height)

			resEvents.EmitEvent(sdk.NewEvent("reject-call", sdk.Attribute{Key: "callId", Value: fmt.Sprintf("%d", start)}))
			logger.Info(
				fmt.Sprintf("Removing %d call as not approved in time", callId),
			)
		}
	}

	return resEvents.ABCIEvents()
}
