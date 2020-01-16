package types

import "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName = "vm"
	RouteKey   = ModuleName

	Codespace types.CodespaceType = ModuleName
)

type Contract []byte
