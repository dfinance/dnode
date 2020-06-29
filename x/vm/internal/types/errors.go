// Errors.
package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// VM related status codes
	VMCodeExecuted = 4001
)

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// Msg contract bytes (Module / Script) are empty.
	ErrEmptyContract = sdkErrors.Register(ModuleName, 101, "contract code is empty")
	// Move VM crashes, means don't return response, disconnect, etc (that error breaks consensus)
	ErrVMCrashed = sdkErrors.Register(ModuleName, 102, "VM has crashed / not reachable")

	// Wrong address length.
	ErrWrongAddressLength = sdkErrors.Register(ModuleName, 201, "address passed to vm has wrong length")
	// Converting msg.Args VMTypeTag to string failed.
	ErrWrongArgTypeTag = sdkErrors.Register(ModuleName, 202, "something wrong with argument type")
	// Empty argument value.
	ErrWrongArgValue = sdkErrors.Register(ModuleName, 203, "something wrong with argument value")
	// Move VM can't process request correctly: number of resp.Executions != 1.
	ErrWrongExecutionResponse = sdkErrors.Register(ModuleName, 204, "wrong execution response from vm")

	// Data source: value missed in Data Source server.
	ErrDSMissedValue = sdkErrors.Register(ModuleName, 401, "value is missed in storage")

	// Gov: invalid proposal
	ErrGovInvalidProposal = sdkErrors.Register(ModuleName, 500, "invalid proposal")
)
