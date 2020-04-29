// Basic constants and function to work with types.
package types

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
)

const (
	ModuleName = "vm"

	StoreKey  = ModuleName
	RouterKey = ModuleName

	VmAddressLength = 32
	VmGasPrice      = 1
	VmUnknowTagType = -1
)

// VM related variables.
var (
	KeyGenesis = []byte("gen") // used to save genesis
)

// Type of Move contract (bytes).
type Contract []byte

// Convert bech32 to libra hex.
func Bech32ToLibra(addr types.AccAddress) []byte {
	return append(addr, make([]byte, 4)...)
}

// Convert VMAccessPath to hex string
func PathToHex(path vm_grpc.VMAccessPath) string {
	return fmt.Sprintf("Access path: \n"+
		"\tAddress: %s\n"+
		"\tPath:    %s\n"+
		"\tKey:     %s\n", hex.EncodeToString(path.Address), hex.EncodeToString(path.Path), hex.EncodeToString(common_vm.MakePathKey(path)))
}

// Get TypeTag by string TypeTag representation.
func GetVMTypeByString(typeTag string) (vm_grpc.VMTypeTag, error) {
	if val, ok := vm_grpc.VMTypeTag_value[typeTag]; !ok {
		return VmUnknowTagType, fmt.Errorf("can't find tag type %s, check correctness of type value", typeTag)
	} else {
		return vm_grpc.VMTypeTag(val), nil
	}
}

// Convert TypeTag to string representation.
func VMTypeToString(tag vm_grpc.VMTypeTag) (string, error) {
	if val, ok := vm_grpc.VMTypeTag_name[int32(tag)]; !ok {
		return "", fmt.Errorf("can't find string representation of type %d, check correctness of type value", tag)
	} else {
		return val, nil
	}
}

// Convert TypeTag to string representation with panic.
func VMTypeToStringPanic(tag vm_grpc.VMTypeTag) string {
	if val, ok := vm_grpc.VMTypeTag_name[int32(tag)]; !ok {
		panic(fmt.Errorf("can't find string representation of type %d, check correctness of type value", tag))
	} else {
		return val
	}
}
