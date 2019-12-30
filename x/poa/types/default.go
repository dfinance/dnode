package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName                          = "poa"
	RouterKey                           = ModuleName
	DefaultCodespace  sdk.CodespaceType = ModuleName
	DefaultParamspace                   = ModuleName
)

var (
	ValidatorsCountKey = []byte("validators_count")
	ValidatorsListKey  = []byte("validators")
)

type GenesisState struct {
	Parameters    Params     `json:"parameters"`
	PoAValidators Validators `json:"validators"`
}
