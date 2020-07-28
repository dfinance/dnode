package v0_7

import (
	"fmt"

	v06 "github.com/dfinance/dnode/x/multisig/internal/legacy/v0_6"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Migrate migrates v0.6 module state to v0.7 version.
// - Params.IntervalToExecute changed to a new default value;
// - Call.Failed field removed (using call.Error now);
// - Queue items blockHeight is reset to 0;
func Migrate(oldState v06.GenesisState) (GenesisState, error) {
	newState := GenesisState{
		Parameters: v06.Params{
			IntervalToExecute: types.DefIntervalToExecute,
		},
		LastCallID: oldState.LastCallID,
		QueueItems: make([]v06.GenesisQueueItem, 0, len(oldState.QueueItems)),
		CallItems:  make([]GenesisCallItem, 0, len(oldState.CallItems)),
	}

	for _, oldItem := range oldState.QueueItems {
		newItem := v06.GenesisQueueItem{
			CallID:      oldItem.CallID,
			BlockHeight: 0,
		}

		newState.QueueItems = append(newState.QueueItems, newItem)
	}

	for _, oldCallItem := range oldState.CallItems {
		newCallItem := GenesisCallItem{
			Call: Call{
				ID:       oldCallItem.Call.ID,
				UniqueID: oldCallItem.Call.UniqueID,
				Creator:  oldCallItem.Call.Creator,
				Approved: oldCallItem.Call.Approved,
				Executed: oldCallItem.Call.Executed,
				Rejected: oldCallItem.Call.Rejected,
				Msg:      oldCallItem.Call.Msg,
				MsgRoute: oldCallItem.Call.MsgRoute,
				MsgType:  oldCallItem.Call.MsgType,
				Height:   oldCallItem.Call.Height,
			},
			Votes: oldCallItem.Votes,
		}
		if oldCallItem.Call.Failed {
			newCallItem.Call.Error = fmt.Sprintf("failed: %v", oldCallItem.Call.Error)
		}

		newState.CallItems = append(newState.CallItems, newCallItem)
	}

	return newState, nil
}
