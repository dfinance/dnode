package common_vm

import (
	"bytes"
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	// Default address length (Move address length)
	VMAddressLength = 20
)

var (
	// Storage keys
	KeyDelimiter = []byte(":") // we should rely on this delimiter (for bytes.Split for example) as VM accessPath.Path might include symbols like: [':', '@',..]
	VMKey        = []byte("vm")
	// Move stdlib addresses
	StdLibAddress         = make([]byte, VMAddressLength)
	StdLibAddressShortStr = "0x1"
)

// DSDataMiddleware defines prototype for DataSource server middleware.
type DSDataMiddleware func(ctx sdk.Context, path *vm_grpc.VMAccessPath) ([]byte, error)

// VMStorage interface used by other keepers to get/set VM data.
type VMStorage interface {
	// VM accessPath for an oracle asset price resource
	GetOracleAccessPath(assetCode dnTypes.AssetCode) *vm_grpc.VMAccessPath

	// Setters / getters for a VM storage values
	SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte)
	GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte

	// Delete VM value from a VM storage
	DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath)

	// Check value in a VM storage exists
	HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool
}

// GetPathKey returns storage key for VM values from VM AccessPath.
func GetPathKey(path *vm_grpc.VMAccessPath) []byte {
	return bytes.Join(
		[][]byte{
			VMKey,
			path.Address,
			path.Path,
		},
		KeyDelimiter,
	)
}

// GetPathPrefixKey returns strage key prefix for VM values (used for iteration).
func GetPathPrefixKey() []byte {
	return append(VMKey, KeyDelimiter...)
}

// MustParsePathKey parses VM storage key and panics on failure.
func MustParsePathKey(key []byte) *vm_grpc.VMAccessPath {
	accessPath := vm_grpc.VMAccessPath{}

	// we expect key to be correct: vm:{address_20bytes}:{path_at_least_1byte}
	expectedMinLen := len(VMKey) + len(KeyDelimiter) + VMAddressLength + len(KeyDelimiter) + 1
	if len(key) < expectedMinLen {
		panic(fmt.Errorf("key %q: invalid length: min expected: %d", string(key), expectedMinLen))
	}

	// calc indices (end index is the next one of the real end idx)
	prefixStartIdx := 0
	prefixEndIdx := prefixStartIdx + len(VMKey)
	delimiterFirstStartIdx := prefixEndIdx
	delimiterFirstEndIdx := delimiterFirstStartIdx + len(KeyDelimiter)
	addressStartIdx := delimiterFirstEndIdx
	addressEndIdx := addressStartIdx + VMAddressLength
	delimiterSecondStartIdx := addressEndIdx
	delimiterSecondEndIdx := delimiterSecondStartIdx + len(KeyDelimiter)
	pathStartIdx := delimiterSecondEndIdx

	// split key
	prefixValue := key[prefixStartIdx:prefixEndIdx]
	delimiterFirstValue := key[delimiterFirstStartIdx:delimiterFirstEndIdx]
	addressValue := key[addressStartIdx:addressEndIdx]
	delimiterSecondValue := key[delimiterSecondStartIdx:delimiterSecondEndIdx]
	pathValue := key[pathStartIdx:]

	// validate
	if !bytes.Equal(prefixValue, VMKey) {
		panic(fmt.Errorf("key %q: prefix: invalid", string(key)))
	}
	if !bytes.Equal(delimiterFirstValue, KeyDelimiter) {
		panic(fmt.Errorf("key %q: 1st delimiter: invalid", string(key)))
	}
	if !bytes.Equal(delimiterSecondValue, KeyDelimiter) {
		panic(fmt.Errorf("key %q: 2nd delimiter: invalid", string(key)))
	}

	accessPath.Address = addressValue
	accessPath.Path = pathValue

	return &accessPath
}

// Bech32ToLibra converts Bech32 to Libra hex.
func Bech32ToLibra(addr sdk.AccAddress) []byte {
	return addr.Bytes()
}

func init() {
	StdLibAddress[VMAddressLength-1] = 1
}
