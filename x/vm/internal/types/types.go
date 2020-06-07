// Basic constants and function to work with types.
package types

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/common_vm"
)

const (
	ModuleName = "vm"

	StoreKey  = ModuleName
	RouterKey = ModuleName

	VmGasPrice      = 1
	VmUnknowTagType = -1
)

// VM related variables.
var (
	KeyGenesis = []byte("gen") // used to save genesis
)

// Type of Move contract (bytes).
type Contract []byte

// Convert VMAccessPath to hex string
func PathToHex(path *vm_grpc.VMAccessPath) string {
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

// VMWriteOp to string.
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

// Stake trace.

// Writeset to string.
func WriteSetToString(value *vm_grpc.VMValue) string {
	return fmt.Sprintf("\t%s: \n"+
		"\t\tAddress: %s\n"+
		"\t\tPath: %s\n"+
		"\t\tValue: %s\n",
		VMWriteOpToString(value.Type), hex.EncodeToString(value.Path.Address),
		hex.EncodeToString(value.Path.Path), hex.EncodeToString(value.Value),
	)
}

// Contract status.
func ExecStatusToString(status vm_grpc.ContractStatus, sstruct *vm_grpc.VMStatus) string {
	return fmt.Sprintf("Status %s: \n"+
		"\tMajor code: %d\n"+
		"\tStr status: %s\n"+
		"\tSub status: %d\n"+
		"\tMesage: %s\n", status.String(), sstruct.MajorStatus, GetStrCode(strconv.FormatUint(sstruct.MajorStatus, 10)), sstruct.SubStatus, sstruct.Message)
}

// Event to string.
func EventToString(event *vm_grpc.VMEvent) string {
	return fmt.Sprintf("\tType: %s\n"+
		"\t\tKey: %s\n"+
		"\t\tSequence number: %d\n"+
		"\t\tValue: %s\n",
		VMTypeToStringPanic(event.Type.Tag), hex.EncodeToString(event.Key),
		event.SequenceNumber, hex.EncodeToString(event.EventData))
}

// Print VM stack trace if contract is not executed successful.
func PrintVMStackTrace(txId []byte, log log.Logger, exec *vm_grpc.VMExecuteResponse) {
	stackTrace := fmt.Sprintf("Stack trace %X: \n", txId)

	// print common status
	stackTrace += ExecStatusToString(exec.Status, exec.StatusStruct)
	stackTrace += "Events: \n"

	if len(exec.Events) == 0 {
		stackTrace += "\tno events\n"
	}

	for _, event := range exec.Events {
		stackTrace += EventToString(event)
	}

	// print all write sets
	stackTrace += "Write set: \n"

	if len(exec.WriteSet) == 0 {
		stackTrace += "\tempty writeset\n"
	}

	for _, ws := range exec.WriteSet {
		stackTrace += WriteSetToString(ws)
	}

	log.Debug(stackTrace)
}
