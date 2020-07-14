package types

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

const (
	// Error codes in json.
	jsonErrorCodes = `{
		"3001": "MALFORMED",
		"2007": "LOCAL_REFERENCE_ERROR",
		"3010": "DUPLICATE_TABLE",
		"1": "INVALID_SIGNATURE",
		"3018": "BAD_U32",
		"1047": "PACK_TYPE_MISMATCH_ERROR",
		"1051": "READREF_EXISTS_MUTABLE_BORROW_ERROR",
		"1022": "POP_REFERENCE_ERROR",
		"1053": "WRITEREF_RESOURCE_ERROR",
		"1060": "BORROWGLOBAL_TYPE_MISMATCH_ERROR",
		"1065": "MOVETOSENDER_NO_RESOURCE_ERROR",
		"4000": "UNKNOWN_RUNTIME_STATUS",
		"1006": "INVALID_RESOURCE_FIELD",
		"4002": "OUT_OF_GAS",
		"18": "SENDING_ACCOUNT_FROZEN",
		"4004": "RESOURCE_ALREADY_EXISTS",
		"1005": "RECURSIVE_STRUCT_DEFINITION",
		"1025": "BR_TYPE_MISMATCH_ERROR",
		"2008": "STORAGE_ERROR",
		"16": "GAS_UNIT_PRICE_ABOVE_MAX_BOUND",
		"4024": "CREATE_NULL_ACCOUNT",
		"4022": "NATIVE_FUNCTION_ERROR",
		"1055": "WRITEREF_NO_MUTABLE_REFERENCE_ERROR",
		"1027": "STLOC_TYPE_MISMATCH_ERROR",
		"1001": "INDEX_OUT_OF_BOUNDS",
		"18446744073709551615": "UNKNOWN_STATUS",
		"10": "EXCEEDED_MAX_TRANSACTION_SIZE",
		"1064": "MOVETOSENDER_TYPE_MISMATCH_ERROR",
		"3003": "UNKNOWN_VERSION",
		"1033": "FREEZEREF_EXISTS_MUTABLE_BORROW_ERROR",
		"2001": "OUT_OF_BOUNDS_INDEX",
		"2005": "PC_OVERFLOW",
		"1044": "BORROWLOC_EXISTS_BORROW_ERROR",
		"4009": "DATA_FORMAT_ERROR",
		"1015": "INCONSISTENT_FIELDS",
		"3020": "BAD_U128",
		"1007": "INVALID_FALL_THROUGH",
		"1076": "NUMBER_OF_TYPE_ARGUMENTS_MISMATCH",
		"1071": "EXTRANEOUS_ACQUIRES_RESOURCE_ANNOTATION_ERROR",
		"14": "MAX_GAS_UNITS_BELOW_MIN_TRANSACTION_GAS_UNITS",
		"1052": "WRITEREF_TYPE_MISMATCH_ERROR",
		"1074": "GLOBAL_REFERENCE_ERROR",
		"17": "INVALID_GAS_SPECIFIER",
		"1086": "INVALID_LOOP_BREAK",
		"3009": "UNEXPECTED_SIGNATURE_TYPE",
		"1054": "WRITEREF_EXISTS_BORROW_ERROR",
		"1010": "UNBALANCED_STACK",
		"1014": "UNIMPLEMENTED_HANDLE",
		"6": "TRANSACTION_EXPIRED",
		"2003": "EMPTY_VALUE_STACK",
		"3007": "UNKNOWN_OPCODE",
		"3016": "BAD_ULEB_U32",
		"1011": "INVALID_MAIN_FUNCTION_SIGNATURE",
		"1040": "MOVELOC_UNAVAILABLE_ERROR",
		"1059": "EXISTS_RESOURCE_TYPE_MISMATCH_ERROR",
		"3019": "BAD_U64",
		"7": "SENDING_ACCOUNT_DOES_NOT_EXIST",
		"1029": "UNSAFE_RET_LOCAL_OR_RESOURCE_STILL_BORROWED",
		"1070": "MISSING_ACQUIRES_RESOURCE_ANNOTATION_ERROR",
		"4005": "EVICTED_ACCOUNT_ACCESS",
		"1017": "LOOKUP_FAILED",
		"1084": "EMPTY_CODE_UNIT",
		"4017": "ARITHMETIC_ERROR",
		"4018": "DYNAMIC_REFERENCE_ERROR",
		"2012": "VM_STARTUP_FAILURE",
		"1089": "TOO_MANY_LOCALS",
		"1087": "INVALID_LOOP_CONTINUE",
		"8": "REJECTED_WRITE_SET",
		"3013": "UNKNOWN_KIND",
		"1000": "UNKNOWN_VERIFICATION_ERROR",
		"1085": "INVALID_LOOP_SPLIT",
		"4019": "CODE_DESERIALIZATION_ERROR",
		"1021": "MISSING_DEPENDENCY",
		"1083": "MALFORMED_CONSTANT_DATA",
		"1050": "READREF_RESOURCE_ERROR",
		"1041": "MOVELOC_EXISTS_BORROW_ERROR",
		"4020": "EXECUTION_STACK_OVERFLOW",
		"1008": "JOIN_FAILURE",
		"1081": "LINKER_ERROR",
		"11": "UNKNOWN_SCRIPT",
		"1034": "BORROWFIELD_TYPE_MISMATCH_ERROR",
		"1031": "RET_BORROWED_MUTABLE_REFERENCE_ERROR",
		"12": "UNKNOWN_MODULE",
		"1009": "NEGATIVE_STACK_SIZE_WITHIN_BLOCK",
		"1032": "FREEZEREF_TYPE_MISMATCH_ERROR",
		"1036": "BORROWFIELD_EXISTS_MUTABLE_BORROW_ERROR",
		"1019": "TYPE_RESOLUTION_FAILURE",
		"1037": "COPYLOC_UNAVAILABLE_ERROR",
		"1045": "CALL_TYPE_MISMATCH_ERROR",
		"1057": "BOOLEAN_OP_TYPE_MISMATCH_ERROR",
		"2002": "OUT_OF_BOUNDS_RANGE",
		"1046": "CALL_BORROWED_MUTABLE_REFERENCE_ERROR",
		"4016": "ABORTED",
		"3004": "UNKNOWN_TABLE_TYPE",
		"2000": "UNKNOWN_INVARIANT_VIOLATION_ERROR",
		"20": "CURRENCY_INFO_DOES_NOT_EXIST",
		"2010": "EVENT_KEY_MISMATCH",
		"3005": "UNKNOWN_SIGNATURE_TYPE",
		"4014": "VALUE_DESERIALIZATION_ERROR",
		"1023": "POP_RESOURCE_ERROR",
		"1028": "STLOC_UNSAFE_TO_DESTROY_ERROR",
		"1078": "UNUSED_LOCALS_SIGNATURE",
		"1066": "CREATEACCOUNT_TYPE_MISMATCH_ERROR",
		"1043": "BORROWLOC_UNAVAILABLE_ERROR",
		"1049": "READREF_TYPE_MISMATCH_ERROR",
		"3000": "UNKNOWN_BINARY_ERROR",
		"1042": "BORROWLOC_REFERENCE_ERROR",
		"3017": "BAD_U16",
		"4015": "DUPLICATE_MODULE_NAME",
		"1068": "NO_MODULE_HANDLES",
		"4021": "CALL_STACK_OVERFLOW",
		"13": "MAX_GAS_UNITS_EXCEEDS_MAX_GAS_UNITS_BOUND",
		"1069": "POSITIVE_STACK_SIZE_AT_BLOCK_END",
		"4007": "TYPE_ERROR",
		"1067": "MODULE_ADDRESS_DOES_NOT_MATCH_SENDER",
		"3008": "BAD_HEADER_TABLE",
		"1024": "RELEASEREF_TYPE_MISMATCH_ERROR",
		"3014": "UNKNOWN_NATIVE_STRUCT_FLAG",
		"4010": "INVALID_DATA",
		"4012": "CANNOT_WRITE_EXISTING_RESOURCE",
		"4013": "VALUE_SERIALIZATION_ERROR",
		"1030": "RET_TYPE_MISMATCH_ERROR",
		"1016": "UNUSED_FIELD",
		"19": "UNABLE_TO_DESERIALIZE_ACCOUNT",
		"1072": "DUPLICATE_ACQUIRES_RESOURCE_ANNOTATION_ERROR",
		"2004": "EMPTY_CALL_STACK",
		"1039": "COPYLOC_EXISTS_BORROW_ERROR",
		"3002": "BAD_MAGIC",
		"1013": "INVALID_MODULE_HANDLE",
		"2": "INVALID_AUTH_KEY",
		"1075": "CONTRAINT_KIND_MISMATCH",
		"3015": "BAD_ULEB_U16",
		"1026": "ABORT_TYPE_MISMATCH_ERROR",
		"1002": "RANGE_OUT_OF_BOUNDS",
		"1056": "INTEGER_OP_TYPE_MISMATCH_ERROR",
		"9": "INVALID_WRITE_SET",
		"1038": "COPYLOC_RESOURCE_ERROR",
		"1080": "ZERO_SIZED_STRUCT",
		"15": "GAS_UNIT_PRICE_BELOW_MIN_BOUND",
		"4011": "REMOTE_DATA_ERROR",
		"4023": "GAS_SCHEDULE_ERROR",
		"3": "SEQUENCE_NUMBER_TOO_OLD",
		"1063": "MOVEFROM_NO_RESOURCE_ERROR",
		"2014": "INVALID_CODE_CACHE",
		"1058": "EQUALITY_OP_TYPE_MISMATCH_ERROR",
		"1003": "INVALID_SIGNATURE_TOKEN",
		"1004": "INVALID_FIELD_DEF",
		"0": "UNKNOWN_VALIDATION_STATUS",
		"2006": "VERIFICATION_ERROR",
		"1088": "UNSAFE_RET_UNUSED_RESOURCES",
		"1077": "LOOP_IN_INSTANTIATION_GRAPH",
		"4008": "MISSING_DATA",
		"3011": "VERIFIER_INVARIANT_VIOLATION",
		"5": "INSUFFICIENT_BALANCE_FOR_TRANSACTION_FEE",
		"1048": "UNPACK_TYPE_MISMATCH_ERROR",
		"1012": "DUPLICATE_ELEMENT",
		"1073": "INVALID_ACQUIRES_RESOURCE_ANNOTATION_ERROR",
		"2011": "UNREACHABLE",
		"1018": "VISIBILITY_MISMATCH",
		"1082": "INVALID_CONSTANT_TYPE",
		"3006": "UNKNOWN_SERIALIZED_TYPE",
		"4": "SEQUENCE_NUMBER_TOO_NEW",
		"1035": "BORROWFIELD_BAD_FIELD_ERROR",
		"3012": "UNKNOWN_NOMINAL_RESOURCE",
		"4001": "EXECUTED",
		"4003": "RESOURCE_DOES_NOT_EXIST",
		"4006": "ACCOUNT_ADDRESS_ALREADY_EXISTS",
		"1020": "TYPE_MISMATCH",
		"1061": "BORROWGLOBAL_NO_RESOURCE_ERROR",
		"1079": "UNUSED_TYPE_SIGNATURE",
		"2009": "INTERNAL_TYPE_ERROR",
		"2013": "NATIVE_FUNCTION_INTERNAL_INCONSISTENCY",
		"1062": "MOVEFROM_TYPE_MISMATCH_ERROR"
	}`

	// VM error is unknown.
	VMErrUnknown = "unknown"
)

var (
	// VM errors list.
	errorCodes map[string]string
)

// Load VM errors.
func init() {
	// File with errors.
	errorCodes = make(map[string]string)
	if err := json.Unmarshal([]byte(jsonErrorCodes), &errorCodes); err != nil {
		panic(err)
	}
}

// Get major code in string representation.
func GetStrCode(majorCode string) string {
	if v, ok := errorCodes[majorCode]; ok {
		return v
	}

	return VMErrUnknown
}

// VM error response.
type VMStatus struct {
	Status    string `json:"status"`               // Status of error: error/discard.
	MajorCode string `json:"major_code,omitempty"` // Major code.
	SubCode   string `json:"sub_code,omitempty"`   // Sub code.
	StrCode   string `json:"str_code,omitempty"`   // Detailed exaplantion of code.
	Message   string `json:"message,omitempty"`    // Message.
}

// Create new vm error.
func NewVMStatus(status, majorCode, subCode, message string) VMStatus {
	strCode := ""

	if status != AttributeValueStatusKeep {
		strCode = GetStrCode(majorCode)
	}

	return VMStatus{
		Status:    status,
		MajorCode: majorCode,
		SubCode:   subCode,
		Message:   message,
		StrCode:   strCode,
	}
}

// VM error as string.
func (status VMStatus) String() string {
	return fmt.Sprintf("VM status:"+
		"\tStatus: %s\n"+
		"\tMajor code: %s\n"+
		"\tString code: %s\n"+
		"\tSub code: %s\n"+
		"\tMessage:  %s\n",
		status.Status, status.MajorCode, status.StrCode,
		status.SubCode, status.Message,
	)
}

// VM error responses.
type VMStatuses []VMStatus

// VM error responses to string.
func (statuses VMStatuses) String() string {
	s := ""

	for _, status := range statuses {
		s += status.String()
	}

	return s
}

// Response contains tx hash with vm errors.
type TxVMStatus struct {
	Hash       string     `json:"hash"`
	VMStatuses VMStatuses `json:"vm_status"`
}

// New Tx VM response.
func NewTxVMStatus(hash string, statuses VMStatuses) TxVMStatus {
	return TxVMStatus{
		Hash:       hash,
		VMStatuses: statuses,
	}
}

// TxVMResponse to string.
func (tx TxVMStatus) String() string {
	return fmt.Sprintf("Tx:"+
		"\tHash: %s\n"+
		"\tStatuses: %s\n",
		tx.Hash, tx.VMStatuses.String())
}

// New VM error from events.
func NewVMStatusFromABCILogs(tx types.TxResponse) TxVMStatus {
	statuses := make(VMStatuses, 0)

	for _, log := range tx.Logs {
		for _, event := range log.Events {
			isFound := false

			if event.Type == EventTypeContractStatus {
				status := ""
				majorCode := ""
				subCode := ""
				message := ""

				for _, attr := range event.Attributes {
					// find that it's event contains contract status.
					if attr.Key == AttributeStatus {
						status = attr.Value

						if status == AttributeValueStatusDiscard || status == AttributeValueStatusError {
							isFound = true
							break
						}
					}
				}

				// event found.
				if isFound {
					for _, attr := range event.Attributes {
						switch attr.Key {
						case AttributeErrMajorStatus:
							majorCode = attr.Value

						case AttributeErrSubStatus:
							subCode = attr.Value

						case AttributeErrMessage:
							message = attr.Value
						}
					}
				}

				statuses = append(statuses, NewVMStatus(status, majorCode, subCode, message))
			}
		}
	}

	return NewTxVMStatus(tx.TxHash, statuses)
}
