package types

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/types"
	vm "wings-blockchain/x/core/protos"
)

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

var (
	VMModuleType = []byte("module")
	KeyDelimiter = []byte(":")

	zeroBytes = make([]byte, 12)
)

func MakePathKey(path vm.VMAccessPath, resourceType []byte) []byte {
	return bytes.Join(
		[][]byte{
			path.Address,
			[]byte(resourceType),
			path.Path,
		},
		KeyDelimiter,
	)
}

func EncodeAddress(address types.AccAddress) []byte {
	return append(address, zeroBytes...)
}

func DecodeAddress(address []byte) types.AccAddress {
	return address[:types.AddrLen]
}
