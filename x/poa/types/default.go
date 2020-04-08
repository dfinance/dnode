// Described types for PoA module.
package types

const (
	ModuleName        = "poa"
	RouterKey         = ModuleName
	DefaultParamspace = ModuleName
)

var (
	ValidatorsCountKey = []byte("validators_count") // Count key in DB to count validators.
	ValidatorsListKey  = []byte("validators")       // Key in DB to store validators.
)

// Genesis state parameters contains genesis data.
type GenesisState struct {
	Parameters    Params     `json:"parameters"`
	PoAValidators Validators `json:"validators"`
}
