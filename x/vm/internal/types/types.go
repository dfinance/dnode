package types

import "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName = "vm"

	StoreKey  = ModuleName
	RouterKey = ModuleName

	Codespace         types.CodespaceType = ModuleName
	DefaultParamspace                     = ModuleName
)

type Contract []byte
type GenesisState struct {
	Parameters Params `json:"parameters"`
}
