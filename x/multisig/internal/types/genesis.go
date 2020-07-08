package types

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	Parameters Params `json:"parameters"`
}

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	params := s.Parameters
	if err := params.Validate(); err != nil {
		return err
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Parameters: NewParams(DefIntervalToExecute),
	}
}
