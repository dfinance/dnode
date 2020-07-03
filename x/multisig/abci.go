package multisig

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/keeper"
)

// EndBlocker processes active, rejected calls and their confirmations.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	logger := k.GetLogger(ctx)
	eventManager := sdk.NewEventManager()

	// define iteration start range
	start := ctx.BlockHeight() - k.GetIntervalToExecute(ctx)
	if start < 0 {
		start = 0
	}

	// iterate over active calls (over queue)
	activeIterator := k.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
	defer activeIterator.Close()

	eventManager.EmitEvent(NewActiveCallsEvent(start))
	for ; activeIterator.Valid(); activeIterator.Next() {
		bz := activeIterator.Value()

		var callID dnTypes.ID
		ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &callID)

		confirmations, err := k.GetConfirmationsCount(ctx, callID)
		if err != nil {
			//if types.ErrWrongCallId.Is(err) {
			//	continue
			//}

			panic(fmt.Errorf("getting active call %s confirmations: %v", callID.String(), err))
		}

		// check if call is confirmed enough
		if uint16(confirmations) >= k.GetPoaMinConfirmationsCount(ctx) {
			// call confirmed -> execute
			eventManager.EmitEvent(NewExecuteCallEvent(callID))
			call, err := k.GetCall(ctx, callID)
			if err != nil {
				panic(fmt.Errorf("getting active call %s: %v", call.ID.String(), err))
			}
			call.Approved = true

			handler := k.GetRouteHandler(call.Msg.Route())
			if handler == nil {
				panic(fmt.Errorf("handler for route %q: not found", call.Msg.Route()))
			}

			cacheCtx, writeCache := ctx.CacheContext()
			if err := handler(cacheCtx, call.Msg); err != nil {
				// call execution failed, update call status
				call.Failed = true
				call.Error = err.Error()

				eventManager.EmitEvent(NewFailedCallEvent(callID))
				logger.Info(fmt.Sprintf("Call %s execution failed, marking as failed: %v", callID.String(), err))
			} else {
				// call executed
				call.Executed = true
				writeCache()

				eventManager.EmitEvent(NewExecutedCallEvent(callID))
				logger.Info(fmt.Sprintf("Call %s executed, marking as executed", callID.String()))
			}

			// update call and remove from the queue
			k.StoreCall(ctx, call)
			k.RemoveCallFromQueue(ctx, callID, call.Height)
		}
	}

	// iterate over calls that weren't confirmed after the max interval
	if start > k.GetIntervalToExecute(ctx) {
		eventManager.EmitEvent(NewRejectedCallsEvent(start))
		rejectedIterator := k.GetQueueIteratorTill(ctx, start)
		defer rejectedIterator.Close()
		for ; rejectedIterator.Valid(); rejectedIterator.Next() {
			bz := rejectedIterator.Value()

			var callID dnTypes.ID
			ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &callID)

			call, err := k.GetCall(ctx, callID)
			if err != nil {
				panic(fmt.Errorf("getting rejected call %q: %v", call.ID.String(), err))
			}
			call.Rejected = true

			// update call and remove from the queue
			k.StoreCall(ctx, call)
			k.RemoveCallFromQueue(ctx, callID, call.Height)

			eventManager.EmitEvent(NewRejectedCallEvent(callID))
			logger.Info(fmt.Sprintf("Call %s was not approved in time, marking as rejected", callID.String()))
		}
	}

	return []abci.ValidatorUpdate{}
}
