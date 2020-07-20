package types

import (
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	Parameters Params             `json:"parameters" yaml:"parameters"`
	LastCallID *dnTypes.ID        `json:"last_call_id" yaml:"last_call_id"`
	CallItems  []GenesisCallItem  `json:"call_items" yaml:"call_items"`
	QueueItems []GenesisQueueItem `json:"queue_items" yaml:"queue_items"`
}

// GenesisCallItem stores calls info for genesisState.
type GenesisCallItem struct {
	Call  Call  `json:"call" yaml:"call"`
	Votes Votes `json:"votes" yaml:"votes"`
}

// GenesisQueueItem stores calls queue info for genesisState.
type GenesisQueueItem struct {
	CallID      dnTypes.ID `json:"call_id" yaml:"call_id"`
	BlockHeight int64      `json:"block_height" yaml:"block_height"`
}

// Valid checks that genesis state is valid.
// {curBlockHeight} == -1, no blockHeight checks performed.
func (s GenesisState) Validate(curBlockHeight int64) error {
	if err := s.Parameters.Validate(); err != nil {
		return fmt.Errorf("parameters: %w", err)
	}

	callsSet := make(map[string]bool, len(s.CallItems))
	maxCallID := dnTypes.NewZeroID()
	for i, item := range s.CallItems {
		if callsSet[item.Call.ID.String()] {
			return fmt.Errorf("calls[%d]: call_id %q duplicated", i, item.Call.ID.String())
		}

		if err := item.Call.Valid(curBlockHeight); err != nil {
			return fmt.Errorf("calls[%d]: %w", i, err)
		}

		callsSet[item.Call.ID.String()] = true
		if item.Call.ID.GT(maxCallID) {
			maxCallID = item.Call.ID
		}
	}

	for i, item := range s.QueueItems {
		if err := item.CallID.Valid(); err != nil {
			return fmt.Errorf("queue_items[%d]: call_id: %w", i, err)
		}

		if item.BlockHeight < 0 {
			return fmt.Errorf("queue_items[%d]: block_height: LT 0", i)
		}
		if curBlockHeight != -1 && item.BlockHeight > curBlockHeight {
			return fmt.Errorf("queue_items[%d]: block_height: GT current blockHeight", i)
		}

		if !callsSet[item.CallID.String()] {
			return fmt.Errorf("queue_items[%d]: call_id %q not found in the genesisState", i, item.CallID.String())
		}
	}

	if s.LastCallID == nil && len(s.CallItems) != 0 {
		return fmt.Errorf("last_call_id: nil with existing calls")
	}
	if s.LastCallID != nil {
		if err := s.LastCallID.Valid(); err != nil {
			return fmt.Errorf("last_call_id: %w", err)
		}
		if !s.LastCallID.Equal(maxCallID) {
			return fmt.Errorf("last_call_id: not equal to max call ID")
		}
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Parameters: NewParams(DefIntervalToExecute),
		LastCallID: nil,
		CallItems:  make([]GenesisCallItem, 0),
		QueueItems: make([]GenesisQueueItem, 0),
	}
}
