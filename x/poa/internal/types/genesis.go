package types

import (
	"fmt"
)

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	Parameters Params     `json:"parameters" yaml:"parameters"`
	Validators Validators `json:"validators" yaml:"validators"`
}

// Validate checks that genesis state is valid.
// {skipCountValidation} == true is used by genesis Tx, as count should be checked only on chain start.
func (s GenesisState) Validate(skipCountValidation bool) error {
	if err := s.Parameters.Validate(); err != nil {
		return fmt.Errorf("parameters: %w", err)
	}

	validatorsSet := make(map[string]bool, len(s.Validators))
	for i, v := range s.Validators {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validator [%d]: %w", i, err)
		}

		if validatorsSet[v.Address.String()] {
			return fmt.Errorf("validator [%d]: is a duplicate", i)
		}
		validatorsSet[v.Address.String()] = true
	}

	if !skipCountValidation {
		validatorsCount := len(s.Validators)
		if len(s.Validators) < int(s.Parameters.MinValidators) {
			return fmt.Errorf("invalid validators amount: %d should be >= %d", validatorsCount, s.Parameters.MinValidators)
		}
		if len(s.Validators) > int(s.Parameters.MaxValidators) {
			return fmt.Errorf("invalid validators amount: %d should be <= %d", validatorsCount, s.Parameters.MaxValidators)
		}
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Parameters: DefaultParams(),
		Validators: make(Validators, 0),
	}
}
