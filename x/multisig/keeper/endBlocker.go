package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	types "wings-blockchain/x/multisig/types"
	"wings-blockchain/x/poa"
	"fmt"
)

func EndBlocker(ctx sdk.Context, keeper Keeper, poaKeeper poa.Keeper) sdk.Tags {
	logger  := keeper.getLogger(ctx)
	resTags := sdk.NewTags()

	start := ctx.BlockHeight() - types.IntervalToExecute

	if start < 0 {
		start = 0
	}

	// Iterate active calls
	activeIterator := keeper.GetQueueIteratorFromEnd(ctx, start, ctx.BlockHeight())
	defer activeIterator.Close()

	resTags = resTags.AppendTag("start-active-calls-ex", fmt.Sprintf("%d", start))
	for ; activeIterator.Valid(); activeIterator.Next() {
		bs := activeIterator.Value()

		var callId uint64
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bs, &callId)


		confirmations, err := keeper.GetConfirmations(ctx, callId)

		if err != nil {
			panic(err)
		}

		// check if call is confirmed enough
		if uint16(confirmations) >= poaKeeper.GetEnoughConfirmations(ctx) {
			resTags = resTags.AppendTag("execute-call", fmt.Sprintf("%d", callId))

			// call confirmed - execute
			call := keeper.getCallById(ctx, callId)
			call.Approved = true

			handler := keeper.router.GetRoute(call.Msg.Route())

			cacheCtx, writeCache := ctx.CacheContext()
			err := handler(cacheCtx, call.Msg)

			if err == nil {
				// call execution failed, write it to status
				call.Failed = true
				call.Error = err.Error()

				resTags = resTags.AppendTag("failed", fmt.Sprintf("%d", callId))

				logger.Info(
					fmt.Sprintf("Failed execution of %d call, error: %s, marked as failed",
						callId, err.Error()),
				)
			} else {
				call.Executed = true
				writeCache()

				resTags = resTags.AppendTag("executed", fmt.Sprintf("%d", callId))

				logger.Info(
					fmt.Sprintf("Call %d executed completed", callId),
				)
			}

			// save call as executed
			keeper.saveCallById(ctx, callId, call)

			// remove proposal from queue
			keeper.removeCallFromQueue(ctx, callId, call.GetHeight())
		}
	}

	if start > types.IntervalToExecute {
		resTags = resTags.AppendTag("start-rejected-calls-rem", fmt.Sprintf("%d", start))

		// Remove not confirmed calls during intervals
		rejectedIterator := keeper.GetQueueIteratorTill(ctx, start)
		for ; rejectedIterator.Valid(); rejectedIterator.Next() {
			bs := rejectedIterator.Value()

			var callId uint64
			keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bs, &callId)

			call := keeper.getCallById(ctx, callId)
			call.Rejected = true

			keeper.saveCallById(ctx, callId, call)
			keeper.removeCallFromQueue(ctx, callId, call.GetHeight())

			resTags = resTags.AppendTag("reject-call", fmt.Sprintf("%d", callId))

			logger.Info(
				fmt.Sprintf("Removing %d call as not approved in time", callId),
			)
		}
	}

	return resTags
}