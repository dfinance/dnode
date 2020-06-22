// Basic constants and function to work with types.
package types

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/common_vm"
)

const (
	ModuleName = "vm"

	StoreKey     = ModuleName
	RouterKey    = ModuleName
	GovRouterKey = ModuleName

	VmGasPrice       = 1
	VmUnknownTagType = -1
)

// VM related variables.
var (
	KeyGenesis   = []byte("gen") // used to save genesis
	KeyDelimiter = []byte(":")
)

// Type of Move contract (bytes).
type Contract []byte

// Converts gRPC VMAccessPath to hex string.
func VMPathToHex(path *vm_grpc.VMAccessPath) string {
	if path == nil {
		return "nil"
	}

	return fmt.Sprintf("Access path:\n"+
		"  Address: %s\n"+
		"  Path:    %s\n"+
		"  Key:     %s\n",
		hex.EncodeToString(path.Address),
		hex.EncodeToString(path.Path),
		hex.EncodeToString(common_vm.MakePathKey(path)),
	)
}

// Gets gRPC VMTypeTag by enum string representation.
func GetVMTypeByString(typeTag string) (vm_grpc.VMTypeTag, error) {
	if val, ok := vm_grpc.VMTypeTag_value[typeTag]; !ok {
		return VmUnknownTagType, fmt.Errorf("can't find tag VMTypeTag %s, check correctness of type value", typeTag)
	} else {
		return vm_grpc.VMTypeTag(val), nil
	}
}

// Convert gRPC VMTypeTag to string representation.
func VMTypeTagToString(tag vm_grpc.VMTypeTag) (string, error) {
	if val, ok := vm_grpc.VMTypeTag_name[int32(tag)]; !ok {
		return "", fmt.Errorf("can't find string representation of VMTypeTag %d, check correctness of type value", tag)
	} else {
		return val, nil
	}
}

// Convert gRPC VMTypeTag to string representation, panics on error.
func VMTypeTagToStringPanic(tag vm_grpc.VMTypeTag) string {
	val, err := VMTypeTagToString(tag)
	if err != nil {
		panic(err)
	}

	return val
}

// Convert gRPC LcsTag to string representation (recursive).
// <indentCount> defines number of prefixed indent string for each line.
func VMLCSTagToString(tag *vm_grpc.LcsTag, indentCount ...int) (string, error) {
	const strIndent = "  "

	curIndentCount := 0
	if len(indentCount) > 1 {
		return "", fmt.Errorf("invalid indentCount length")
	}
	if len(indentCount) == 1 {
		curIndentCount = indentCount[0]
	}
	if curIndentCount < 0 {
		return "", fmt.Errorf("invalid indentCount")
	}

	strBuilder := strings.Builder{}

	// Helper funcs
	buildStrIndent := func() string {
		str := ""
		for i := 0; i < curIndentCount; i++ {
			str += strIndent
		}
		return str
	}

	buildErr := func(comment string, err error) error {
		return fmt.Errorf("indent %d: %s: %w", curIndentCount, comment, err)
	}

	buildLcsTypeStr := func(t vm_grpc.LcsType) (string, error) {
		val, ok := vm_grpc.LcsType_name[int32(t)]
		if !ok {
			return "", fmt.Errorf("can't find string representation of LcsTag %d, check correctness of type value", t)
		}
		return val, nil
	}

	// Print current tag with recursive func call for fields
	if tag == nil {
		strBuilder.WriteString("nil")
		return strBuilder.String(), nil
	}

	indentStr := buildStrIndent()
	strBuilder.WriteString("nil\n")

	// Field: TypeTag
	typeTagStr, err := buildLcsTypeStr(tag.TypeTag)
	if err != nil {
		return "", buildErr("TypeTag", err)
	}
	strBuilder.WriteString(fmt.Sprintf("%sTypeTag: %s\n", indentStr, typeTagStr))

	// Field: VectorType
	vectorTypeStr, err := VMLCSTagToString(tag.VectorType, curIndentCount+1)
	if err != nil {
		return "", buildErr("VectorType", err)
	}
	strBuilder.WriteString(fmt.Sprintf("%sVectorType: %s\n", indentStr, vectorTypeStr))

	// Field: StructIdent
	if tag.StructIdent != nil {
		strBuilder.WriteString(fmt.Sprintf("%sStructIdent.Address: %s\n", indentStr, hex.EncodeToString(tag.StructIdent.Address)))
		strBuilder.WriteString(fmt.Sprintf("%sStructIdent.Module: %s\n", indentStr, tag.StructIdent.Module))
		strBuilder.WriteString(fmt.Sprintf("%sStructIdent.Name: %s\n", indentStr, tag.StructIdent.Name))
		if len(tag.StructIdent.TypeParams) > 0 {
			for structParamIdx, structParamTag := range tag.StructIdent.TypeParams {
				structParamTagStr, err := VMLCSTagToString(structParamTag, curIndentCount+1)
				if err != nil {
					return "", buildErr(fmt.Sprintf("StructIdent.TypeParams[%d]", structParamIdx), err)
				}
				strBuilder.WriteString(fmt.Sprintf("%sStructIdent.TypeParams[%d]: %s\n", indentStr, structParamIdx, structParamTagStr))
			}
		} else {
			strBuilder.WriteString(fmt.Sprintf("%sStructIdent.TypeParams: empty\n", indentStr))
		}
	} else {
		strBuilder.WriteString(fmt.Sprintf("%sStructIdent: nil\n", indentStr))
	}

	return strBuilder.String(), nil
}

// Convert gRPC LcsTag to string representation, panics on error.
// <indentCount> defines number of prefixed indent string for each line.
func VMLCSTagToStringPanic(tag *vm_grpc.LcsTag, indentCount ...int) string {
	val, err := VMLCSTagToString(tag, indentCount...)
	if err != nil {
		panic(err)
	}

	return val
}

// Convert gRPC VmWriteOp to string representation.
func VMWriteOpToString(wOp vm_grpc.VmWriteOp) string {
	switch wOp {
	case vm_grpc.VmWriteOp_Value:
		return "write"
	case vm_grpc.VmWriteOp_Deletion:
		return "del"
	default:
		return "unknown"
	}
}

// Convert gRPC VMValue (writeSet) to string representation.
func VMWriteSetToString(value *vm_grpc.VMValue) string {
	if value == nil {
		return "nil"
	}

	return fmt.Sprintf("\nWriteSet %q:\n"+
		"  Address: %s\n"+
		"  Path: %s\n"+
		"  Value: %s\n",
		VMWriteOpToString(value.Type),
		hex.EncodeToString(value.Path.Address),
		hex.EncodeToString(value.Path.Path),
		hex.EncodeToString(value.Value),
	)
}

// Convert gRPC ContractStatus (contract status) to string representation.
func VMExecStatusToString(status vm_grpc.ContractStatus, sstruct *vm_grpc.VMStatus) string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString(fmt.Sprintf("Exec %q status:\n", status.String()))
	if sstruct != nil {
		strBuilder.WriteString(fmt.Sprintf("  Major code: %d\n", sstruct.MajorStatus))
		strBuilder.WriteString(fmt.Sprintf("  Major status: %s\n", GetStrCode(strconv.FormatUint(sstruct.MajorStatus, 10))))
		strBuilder.WriteString(fmt.Sprintf("  Sub code: %d\n", sstruct.SubStatus))
		strBuilder.WriteString(fmt.Sprintf("  Message: %s\n", sstruct.Message))
	} else {
		strBuilder.WriteString("  VMStatus: nil\n")
	}

	return strBuilder.String()
}

// Convert gRPC VMEvent (event) to string representation.
func VMEventToString(event *vm_grpc.VMEvent) string {
	strBuilder := strings.Builder{}

	if event == nil {
		strBuilder.WriteString("nil")
		return strBuilder.String()
	}
	strBuilder.WriteString("\n")

	strBuilder.WriteString("Event\n")
	strBuilder.WriteString(fmt.Sprintf("  SenderAddress: %s\n", hex.EncodeToString(event.SenderAddress)))
	if event.SenderModule != nil {
		strBuilder.WriteString(fmt.Sprintf("  SenderModule.Address: %s\n", hex.EncodeToString(event.SenderModule.Address)))
		strBuilder.WriteString(fmt.Sprintf("  SenderModule.Name: %s\n", event.SenderModule.Name))
	} else {
		strBuilder.WriteString("  SenderModule: nil\n")
	}
	strBuilder.WriteString(fmt.Sprintf("  EventType: %s\n", VMLCSTagToStringPanic(event.EventType, 2)))
	strBuilder.WriteString(fmt.Sprintf("  EventData: %s\n", hex.EncodeToString(event.EventData)))

	return strBuilder.String()
}

// Prints VM stack trace if contract is not executed successfully.
func PrintVMStackTrace(txId []byte, log log.Logger, exec *vm_grpc.VMExecuteResponse) {
	strBuilder := strings.Builder{}

	strBuilder.WriteString(fmt.Sprintf("Stack trace %X:\n", txId))

	// Print common status
	if len(exec.Events) > 0 {
		for eventIdx, event := range exec.Events {
			strBuilder.WriteString(fmt.Sprintf("Events[%d]: %s\n", eventIdx, VMEventToString(event)))
		}
	} else {
		strBuilder.WriteString("Events: empty\n")
	}

	// Print all writeSets
	if len(exec.WriteSet) > 0 {
		for wsIdx, ws := range exec.WriteSet {
			strBuilder.WriteString(fmt.Sprintf("WriteSet[%d]: %s\n", wsIdx, VMWriteSetToString(ws)))
		}
	} else {
		strBuilder.WriteString("WriteSet: empty\n")
	}

	log.Debug(strBuilder.String())
}
