package types

import (
	"fmt"
	"time"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	Issues         []GenesisIssue `json:"issues" yaml:"issues"`
	Withdraws      Withdraws      `json:"withdraws" yaml:"withdraws"`
	LastWithdrawID *dnTypes.ID    `json:"last_withdraw_id" yaml:"last_withdraw_id"`
}

// Valid checks that genesis state is valid.
// Contract: withdraw timestamp check is performed if {curBlockTime} is not empty.
func (s GenesisState) Validate(curBlockTime time.Time) error {
	issueIdsSet := make(map[string]bool, len(s.Issues))
	for i, issue := range s.Issues {
		if err := issue.Valid(); err != nil {
			return fmt.Errorf("issue[%d]: %w", i, err)
		}

		if issueIdsSet[issue.ID] {
			return fmt.Errorf("issue[%d]: duplicated ID %q", i, issue.ID)
		}
		issueIdsSet[issue.ID] = true
	}

	maxWithdrawID := dnTypes.NewZeroID()
	withdrawsIdsSet := make(map[string]bool, len(s.Withdraws))
	for i, withdraw := range s.Withdraws {
		if err := withdraw.Valid(curBlockTime); err != nil {
			return fmt.Errorf("withdraw[%d]: %w", i, err)
		}

		if withdrawsIdsSet[withdraw.ID.String()] {
			return fmt.Errorf("withdraw[%d]: duplicated ID %q", i, withdraw.ID.String())
		}
		withdrawsIdsSet[withdraw.ID.String()] = true

		if withdraw.ID.GT(maxWithdrawID) {
			maxWithdrawID = withdraw.ID
		}
	}

	if s.LastWithdrawID == nil && len(s.Withdraws) != 0 {
		return fmt.Errorf("last_withdraw_id: nil with existing withdraws")
	}
	if s.LastWithdrawID != nil && len(s.Withdraws) == 0 {
		return fmt.Errorf("last_withdraw_id: not nil without existing withdraws")
	}
	if s.LastWithdrawID != nil {
		if err := s.LastWithdrawID.Valid(); err != nil {
			return fmt.Errorf("last_withdraw_id: %w", err)
		}
		if !s.LastWithdrawID.Equal(maxWithdrawID) {
			return fmt.Errorf("last_withdraw_id: not equal to max withdraw ID")
		}
	}

	return nil
}

// GenesisIssue stores issue info for genesisState.
type GenesisIssue struct {
	Issue
	ID string `json:"id" yaml:"id"`
}

func (issue GenesisIssue) Valid() error {
	if err := issue.Issue.Valid(); err != nil {
		return err
	}

	if issue.ID == "" {
		return fmt.Errorf("id: empty")
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{
		LastWithdrawID: nil,
		Issues:         make([]GenesisIssue, 0),
		Withdraws:      Withdraws{},
	}
}
