package types

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

const (
	ModuleName = "vm"

	StoreKey  = ModuleName
	RouterKey = ModuleName

	Codespace         types.CodespaceType = ModuleName
	DefaultParamspace                     = ModuleName

	VmAddressLength = 32
	VmGasPrice      = 1
)

type Contract []byte

var (
	VMModuleType = []byte("module")
	KeyDelimiter = []byte(":")

	zeroBytes = make([]byte, 12)
)

func MakePathKey(path vm_grpc.VMAccessPath, resourceType []byte) []byte {
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
