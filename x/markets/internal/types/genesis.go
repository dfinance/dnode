package types

import (
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Module genesis state object.
type GenesisState struct {
	Markets      Markets     `json:"markets" yaml:"markets"`
	LastMarketID *dnTypes.ID `json:"last_market_id" yaml:"last_market_id"`
}

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	maxMarketID := dnTypes.NewZeroID()
	marketsSet := make(map[string]bool, len(s.Markets))
	for i, m := range s.Markets {
		if err := m.Valid(); err != nil {
			return fmt.Errorf("market[%d]: %v", i, err)
		}

		if marketsSet[m.ID.String()] {
			return fmt.Errorf("market[%d]: duplicated ID", i)
		}
		marketsSet[m.ID.String()] = true

		if m.ID.GT(maxMarketID) {
			maxMarketID = m.ID
		}
	}

	if s.LastMarketID == nil && len(s.Markets) != 0 {
		return fmt.Errorf("last_market_id: nil with existing markets")
	}
	if s.LastMarketID != nil && len(s.Markets) == 0 {
		return fmt.Errorf("last_market_id: not nil without existing markets")
	}
	if s.LastMarketID != nil {
		if err := s.LastMarketID.Valid(); err != nil {
			return fmt.Errorf("last_market_id: %w", err)
		}
		if !s.LastMarketID.Equal(maxMarketID) {
			return fmt.Errorf("last_market_id: not equal to max market ID")
		}
	}

	return nil
}

// DefaultGenesisState returns module default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Markets: Markets{},
	}
}
