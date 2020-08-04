package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// VM related status codes
	VMCodeExecuted = 4001
)

var (
	ErrInternal      = sdkErrors.Register(ModuleName, 100, "internal")
	ErrEmptyContract = sdkErrors.Register(ModuleName, 101, "contract code is empty")
	ErrVMCrashed     = sdkErrors.Register(ModuleName, 102, "VM has crashed / not reachable") // error breaks consensus
	ErrNotFound      = sdkErrors.Register(ModuleName, 103, "not found")

	ErrWrongArgTypeTag        = sdkErrors.Register(ModuleName, 200, "invalid argument type")
	ErrWrongArgValue          = sdkErrors.Register(ModuleName, 201, "invalid argument value")
	ErrWrongExecutionResponse = sdkErrors.Register(ModuleName, 202, "wrong execution response from VM")

	ErrGovInvalidProposal = sdkErrors.Register(ModuleName, 500, "invalid proposal")
)
