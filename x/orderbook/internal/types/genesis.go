package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// GenesisState orderbook state that must be provided at genesis.
type GenesisState struct {
	HistoryItems HistoryItems `json:"history_items" yaml:"history_items"`
}

// Validate checks that genesis state is valid.
func (gs GenesisState) Validate(currentBlockTime time.Time, currentBlockHeight int64) error {
	historyItemIdsSet := make(map[string]bool, len(gs.HistoryItems))
	var maxBlockHeight int64

	getItemId := func(item HistoryItem) string {
		return fmt.Sprintf("%s:%d", item.MarketID, item.BlockHeight)
	}

	for i, item := range gs.HistoryItems {
		if err := item.Valid(); err != nil {
			return fmt.Errorf("historyItem[%d]: %w", i, err)
		}

		if !currentBlockTime.IsZero() && time.Unix(item.Timestamp, 0).After(currentBlockTime) {
			return fmt.Errorf("historyItem[%d]: timestamp after block time", i)
		}

		itemId := getItemId(item)

		if historyItemIdsSet[itemId] {
			return fmt.Errorf("historyItem[%d]: duplicated ID %q", i, itemId)
		}

		historyItemIdsSet[itemId] = true

		if item.BlockHeight > maxBlockHeight {
			maxBlockHeight = item.BlockHeight
		}
	}

	if currentBlockHeight != -1 && maxBlockHeight > currentBlockHeight {
		return fmt.Errorf("historyItem blockHeight: GT current blockHeight")
	}

	return nil
}

// Equal checks whether two gov GenesisState structs are equivalent.
func (gs GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	s1, _ := json.Marshal(gs)
	s2, _ := json.Marshal(data2)
	fmt.Println(string(s1))
	fmt.Println(string(s2))
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// DefaultGenesisState defines default GenesisState for orderbook.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		HistoryItems: HistoryItems{},
	}
}
