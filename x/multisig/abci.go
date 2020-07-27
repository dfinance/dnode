package multisig

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/keeper"
	"github.com/dfinance/dnode/x/poa"
)

// EndBlocker processes active, rejected calls and their confirmations.
func EndBlocker(ctx sdk.Context, msKeeper keeper.Keeper, poaKeeper poa.Keeper) []abci.ValidatorUpdate {
	logger := msKeeper.GetLogger(ctx)
	eventManager := ctx.EventManager()
	prevEventsCnt := len(eventManager.Events())

	// define iteration start range
	start := ctx.BlockHeight() - msKeeper.GetIntervalToExecute(ctx)
	if start < 0 {
		start = 0
	}

	// iterate over active calls (over queue)
	activeIterator := msKeeper.GetQueueIteratorStartEnd(ctx, start, ctx.BlockHeight())
	defer activeIterator.Close()

	for ; activeIterator.Valid(); activeIterator.Next() {
		bz := activeIterator.Value()

		var callID dnTypes.ID
		ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &callID)

		confirmations, err := msKeeper.GetConfirmationsCount(ctx, callID)
		if err != nil {
			//if types.ErrWrongCallId.Is(err) {
			//	continue
			//}

			panic(fmt.Errorf("getting active call %s confirmations: %v", callID.String(), err))
		}

		// check if call is confirmed enough
		if uint16(confirmations) >= poaKeeper.GetEnoughConfirmations(ctx) {
			// call confirmed -> execute
			call, err := msKeeper.GetCall(ctx, callID)
			if err != nil {
				panic(fmt.Errorf("getting active call %s: %v", call.ID.String(), err))
			}
			call.Approved = true
			eventManager.EmitEvent(NewCallStateChangedEvent(callID, AttributeValueApproved))

			handler := msKeeper.GetRouteHandler(call.Msg.Route())
			if handler == nil {
				panic(fmt.Errorf("handler for route %q: not found", call.Msg.Route()))
			}

			cacheCtx, writeCache := ctx.CacheContext()
			if err := handler(cacheCtx, call.Msg); err != nil {
				// call execution failed, update call status
				call.Error = fmt.Sprintf("failed: %v", err.Error())

				eventManager.EmitEvent(NewCallStateChangedEvent(callID, AttributeValueFailed))
				logger.Info(fmt.Sprintf("Call %s execution failed, marking as failed: %v", callID.String(), err))
			} else {
				// call executed
				call.Executed = true

				eventManager.EmitEvents(cacheCtx.EventManager().Events())
				writeCache()

				eventManager.EmitEvent(NewCallStateChangedEvent(callID, AttributeValueExecuted))
				logger.Info(fmt.Sprintf("Call %s executed, marking as executed", callID.String()))
			}

			// update call and remove from the queue
			msKeeper.StoreCall(ctx, call)
			msKeeper.RemoveCallFromQueue(ctx, callID, call.Height)
		}
	}

	// iterate over calls that weren't confirmed after the max interval
	if start > msKeeper.GetIntervalToExecute(ctx) {
		rejectedIterator := msKeeper.GetQueueIteratorTill(ctx, start)
		defer rejectedIterator.Close()
		for ; rejectedIterator.Valid(); rejectedIterator.Next() {
			bz := rejectedIterator.Value()

			var callID dnTypes.ID
			ModuleCdc.MustUnmarshalBinaryLengthPrefixed(bz, &callID)

			call, err := msKeeper.GetCall(ctx, callID)
			if err != nil {
				panic(fmt.Errorf("getting rejected call %q: %v", call.ID.String(), err))
			}
			call.Rejected = true

			// update call and remove from the queue
			msKeeper.StoreCall(ctx, call)
			msKeeper.RemoveCallFromQueue(ctx, callID, call.Height)

			eventManager.EmitEvent(NewCallStateChangedEvent(callID, AttributeValueRejected))
			logger.Info(fmt.Sprintf("Call %s was not approved in time, marking as rejected", callID.String()))
		}
	}

	if curEventsCnt := len(eventManager.Events()); curEventsCnt != prevEventsCnt {
		eventManager.EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))
	}

	return []abci.ValidatorUpdate{}
}
