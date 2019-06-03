package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error codes
const (
	CodeRouteDoesntExist   = 101
	CodeErrWrongCallId     = 102

	CodeErrAlreadyApproved = 201
	CodeErrAlreadyExecuted = 202
	CodeErrNotApproved     = 203

	CodeErrNoVotes		   = 301

	CodeNotValidator       = 401
)

// When msg route doesnt exist
func ErrRouteDoesntExist(route string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeRouteDoesntExist, "route doesn't exists %s", route)
}

// When call with provided id doesnt exist
func ErrWrongCallId(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongCallId, "call %d not found", id)
}


// When no votes found for call
func ErrNoVotes(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNoVotes, "no votes found for call %d", id)
}

// When call already approved by address
func ErrCallAlreadyApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyApproved, "call %d already approved by %s", id, address)
}

// When call not approved by address
func ErrCallNotApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNotApproved, "call %d not approved by %s", id, address)
}

// When call already executed
func ErrAlreadyExecuted(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyExecuted, "call %d already executed", id)
}

func ErrNotValidator(validator string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotValidator, "%s is not a validator", validator)
}
