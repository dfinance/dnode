package types

import (
	"bytes"
	"fmt"
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
	VmUnknowTagType = -1
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
			resourceType,
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

func GetVMTypeByString(typeTag string) (vm_grpc.VMTypeTag, error) {
	if val, ok := vm_grpc.VMTypeTag_value[typeTag]; !ok {
		return VmUnknowTagType, fmt.Errorf("can't find tag type %s, check correctness of type value", typeTag)
	} else {
		return vm_grpc.VMTypeTag(val), nil
	}
}

func VMTypeToString(tag vm_grpc.VMTypeTag) (string, error) {
	if val, ok := vm_grpc.VMTypeTag_name[int32(tag)]; !ok {
		return "", fmt.Errorf("can't find string representation of type %d, check correctness of type value", tag)
	} else {
		return val, nil
	}
}
