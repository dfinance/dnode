package client

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/OneOfOne/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// NewAddressScriptArg convert string to address ScriptTag.
func NewAddressScriptArg(value string) (types.ScriptArg, error) {
	argTypeCode := vm_grpc.VMTypeTag_Address
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	if value == "" {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: empty", value, argTypeName)
	}

	addr, err := sdk.AccAddressFromBech32(value)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	return types.ScriptArg{
		Type:  argTypeCode,
		Value: common_vm.Bech32ToLibra(addr),
	}, nil
}

// NewU8ScriptArg convert string to U8 ScriptTag.
func NewU8ScriptArg(value string) (types.ScriptArg, error) {
	argTypeCode := vm_grpc.VMTypeTag_U8
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	hashParsedValue, err := parseXxHashUint(value)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	uintValue, err := strconv.ParseUint(hashParsedValue, 10, 8)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	return types.ScriptArg{
		Type:  argTypeCode,
		Value: []byte{uint8(uintValue)},
	}, nil
}

// NewU64ScriptArg convert string to U64 ScriptTag.
func NewU64ScriptArg(value string) (types.ScriptArg, error) {
	argTypeCode := vm_grpc.VMTypeTag_U64
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	hashParsedValue, err := parseXxHashUint(value)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	uintValue, err := strconv.ParseUint(hashParsedValue, 10, 64)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}
	argValue := make([]byte, 8)
	binary.LittleEndian.PutUint64(argValue, uintValue)

	return types.ScriptArg{
		Type:  argTypeCode,
		Value: argValue,
	}, nil
}

// NewU128ScriptArg convert string to U128 ScriptTag.
func NewU128ScriptArg(value string) (retTag types.ScriptArg, retErr error) {
	argTypeCode := vm_grpc.VMTypeTag_U128
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	defer func() {
		if recover() != nil {
			retErr = fmt.Errorf("parsing argument %q of type %q: failed", value, argTypeName)
		}
	}()

	hashParsedValue, err := parseXxHashUint(value)
	if err != nil {
		retErr = fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
		return
	}

	bigIntValue := sdk.NewUintFromString(hashParsedValue)
	if bigIntValue.BigInt().BitLen() > 128 {
		retErr = fmt.Errorf("parsing argument %q of type %q: invalid bitLen %d", value, argTypeName, bigIntValue.BigInt().BitLen())
		return
	}

	// BigInt().Bytes() returns BigEndian format, reverse it
	argValue := bigIntValue.BigInt().Bytes()
	for left, right := 0, len(argValue)-1; left < right; left, right = left+1, right-1 {
		argValue[left], argValue[right] = argValue[right], argValue[left]
	}

	// Extend to 16 bytes
	if len(argValue) < 16 {
		zeros := make([]byte, 16-len(argValue))
		argValue = append(argValue, zeros...)
	}

	retTag.Type, retTag.Value = argTypeCode, argValue

	return
}

// NewVectorScriptArg convert string to Vector ScriptTag.
func NewVectorScriptArg(value string) (types.ScriptArg, error) {
	argTypeCode := vm_grpc.VMTypeTag_Vector
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	if value == "" {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: empty", value, argTypeName)
	}

	argValue, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	return types.ScriptArg{
		Type:  argTypeCode,
		Value: argValue,
	}, nil
}

// NewBoolScriptArg convert string to Bool ScriptTag.
func NewBoolScriptArg(value string) (types.ScriptArg, error) {
	argTypeCode := vm_grpc.VMTypeTag_Bool
	argTypeName := vm_grpc.VMTypeTag_name[int32(argTypeCode)]

	valueBool, err := strconv.ParseBool(value)
	if err != nil {
		return types.ScriptArg{}, fmt.Errorf("parsing argument %q of type %q: %w", value, argTypeName, err)
	}

	argValue := []byte{0}
	if valueBool {
		argValue[0] = 1
	}

	return types.ScriptArg{
		Type:  argTypeCode,
		Value: argValue,
	}, nil
}

// parseXxHashUint converts (or skips) xxHash integer format.
func parseXxHashUint(value string) (string, error) {
	if value == "" {
		return "", fmt.Errorf("xxHash parsing: empty")
	}

	if value[0] == '#' {
		seed := xxhash.NewS64(0)
		if len(value) < 2 {
			return "", fmt.Errorf("xxHash parsing: invalid length")
		}

		if _, err := seed.WriteString(strings.ToLower(value[1:])); err != nil {
			return "", fmt.Errorf("xxHash parsing: %w", err)
		}
		value = strconv.FormatUint(seed.Sum64(), 10)
	}

	return value, nil
}

// ConvertStringScriptArguments convert string client argument to ScriptArgs using compiler meta data (arg types).
func ConvertStringScriptArguments(argStrs []string, argTypes []vm_grpc.VMTypeTag) ([]types.ScriptArg, error) {
	if len(argStrs) != len(argTypes) {
		return nil, fmt.Errorf("strArgs / typedArgs length mismatch: %d / %d", len(argStrs), len(argTypes))
	}

	scriptArgs := make([]types.ScriptArg, len(argStrs))
	for argIdx, argStr := range argStrs {
		argType := argTypes[argIdx]
		var scriptArg types.ScriptArg
		var err error

		switch argType {
		case vm_grpc.VMTypeTag_Address:
			scriptArg, err = NewAddressScriptArg(argStr)
		case vm_grpc.VMTypeTag_U8:
			scriptArg, err = NewU8ScriptArg(argStr)
		case vm_grpc.VMTypeTag_U64:
			scriptArg, err = NewU64ScriptArg(argStr)
		case vm_grpc.VMTypeTag_U128:
			scriptArg, err = NewU128ScriptArg(argStr)
		case vm_grpc.VMTypeTag_Bool:
			scriptArg, err = NewBoolScriptArg(argStr)
		case vm_grpc.VMTypeTag_Vector:
			scriptArg, err = NewVectorScriptArg(argStr)
		default:
			return nil, fmt.Errorf("argument[%d]: parsing argument %q: unsupported argType code: %v", argIdx, argStr, argType)
		}

		if err != nil {
			return nil, fmt.Errorf("argument[%d]: %w", argIdx, err)
		}
		scriptArgs[argIdx] = scriptArg
	}

	return scriptArgs, nil
}
