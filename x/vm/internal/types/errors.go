// Errors.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/vm/internal/types/vm_grpc"
)

const (
	CodeEmptyContractCode = 101

	CodeErrWrongAddressLength     = 201
	CodeErrWrongArgTypeTag        = 202
	CodeErrWrongExecutionResponse = 203

	// Errors related to DS (Data Source).
	CodeErrDSMissedValue = 401

	// VM related status codes
	VMCodeExecuted = 4001
)

// Special type for VM crashes, so we can detect later, that it's VM crash error and break consensus
type ErrVMCrashed struct {
	err error
}

// Move VM crashes, means don't return response, disconnect, etc
func NewErrVMCrashed(err error) ErrVMCrashed {
	return ErrVMCrashed{err: err}
}

// Msg contract bytes (Module / Script) are empty
func ErrEmptyContract() sdk.Error {
	return sdk.NewError(Codespace, CodeEmptyContractCode, "contract code is empty, please fill field with compiled contract bytes")
}

// Move VM can't process request correctly: number of resp.Executions != 1
func ErrWrongExecutionResponse(resp vm_grpc.VMExecuteResponses) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongExecutionResponse, "wrong execution response from vm: %v", resp)
}

// Wrong address length
func ErrWrongAddressLength(address sdk.AccAddress) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongAddressLength, "address %s passed to vm has wrong length, it has length %d, but expected %d", address.String(), len(address), VmAddressLength)
}

// Converting msg.Args VMTypeTag to string failed
func ErrWrongArgTypeTag(err error) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongArgTypeTag, "something wrong with argument type: %s", err.Error())
}

// Value missed in Data Source server
func ErrDSMissedValue(accessPath vm_grpc.VMAccessPath) sdk.Error {
	return sdk.NewError(Codespace, CodeErrDSMissedValue, "value is missed in storage: %s", MakePathKey(accessPath))
}
