package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	// validate again knowing current blockHeight
	if err := state.Validate(ctx.BlockHeight()); err != nil {
		panic(err)
	}

	// params
	k.SetIntervalToExecute(ctx, state.Parameters.IntervalToExecute)

	// last callID
	if state.LastCallID != nil {
		k.setLastCallID(ctx, *state.LastCallID)
	}

	// calls, uniqueID-callID matches and votes per call
	for i, callItem := range state.CallItems {
		if !k.router.HasRoute(callItem.Call.Msg.Route()) {
			panic(fmt.Errorf("call[%d]: route %q not registered in keepers router", i, callItem.Call.Msg.Route()))
		}

		k.setCallUniqueIDMatch(ctx, callItem.Call.UniqueID, callItem.Call.ID)
		k.StoreCall(ctx, callItem.Call)
		if len(callItem.Votes) > 0 {
			k.setVotes(ctx, callItem.Call.ID, callItem.Votes)
		}
	}

	// queue
	for _, queueItem := range state.QueueItems {
		k.addCallToQueue(ctx, queueItem.CallID, queueItem.BlockHeight)
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermRead)

	// params and lastCallID
	state := types.GenesisState{
		Parameters: types.Params{
			IntervalToExecute: k.GetIntervalToExecute(ctx),
		},
		LastCallID: k.getLastCallID(ctx),
		CallItems:  make([]types.GenesisCallItem, 0),
		QueueItems: make([]types.GenesisQueueItem, 0),
	}

	// calls with votes
	for _, call := range k.getCalls(ctx) {
		votes, _ := k.GetVotes(ctx, call.ID)
		state.CallItems = append(state.CallItems, types.GenesisCallItem{
			Call:  call,
			Votes: votes,
		})
	}

	// queue
	queueIterator := k.GetQueueIteratorTill(ctx, ctx.BlockHeight())
	defer queueIterator.Close()

	for ; queueIterator.Valid(); queueIterator.Next() {
		callID, blockHeight := types.MustParseQueueKey(queueIterator.Key())
		state.QueueItems = append(state.QueueItems, types.GenesisQueueItem{
			CallID:      callID,
			BlockHeight: blockHeight,
		})
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other module dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
