// Implement error codes and messages.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Error codes.
const (
	CodeErrRouteDoesntExist = 101
	CodeErrWrongCallId      = 102
	CodeErrEmptyRoute       = 103
	CodeErrEmptyType        = 104
	CodeErrOnlyMs           = 105

	CodeErrAlreadyApproved   = 201
	CodeErrAlreadyConfirmed  = 202
	CodeErrAlreadyRerejected = 203
	CodeErrNotApproved       = 204

	CodeErrNoVotes = 301

	CodeNotValidator     = 401
	CodeNotUniqueID      = 402
	CodeNotFoundUniqueID = 403
)

// Only multisig calls supported for module.
func ErrOnlyMultisig(codeSpase sdk.CodespaceType, moduleName string) sdk.Error {
	return sdk.NewError(codeSpase, CodeErrOnlyMs, "module %s does support only multisig calls, see mshandler...", moduleName)
}

// When msg route doesnt exist.
func ErrRouteDoesntExist(route string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrRouteDoesntExist, "route doesn't exists %s", route)
}

// When msg route is empty (could be empty if we use MsMsg interface).
func ErrEmptyRoute(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrEmptyRoute, "msg route is empty for %d call", id)
}

// When msg route is empty (could be empty if we use MsMsg interface).
func ErrEmptyType(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrEmptyType, "msg type is empty for %d call", id)
}

// When call with provided id doesnt exist.
func ErrWrongCallId(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongCallId, "call %d not found", id)
}

// When no votes found for call.
func ErrNoVotes(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNoVotes, "no votes found for call %d", id)
}

// When call already approved by address.
func ErrCallAlreadyApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyApproved, "call %d already approved by %s", id, address)
}

// When call not approved by address.
func ErrCallNotApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNotApproved, "call %d not approved by %s", id, address)
}

// When call already executed.
func ErrAlreadyConfirmed(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyConfirmed, "call %d already confirmed", id)
}

// When tx from not validator.
func ErrNotValidator(validator string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotValidator, "%s is not a validator", validator)
}

// When call already rejected.
func ErrAlreadyRejected(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyRerejected, "%d already rejected", id)
}

// When cant parse call id.
func ErrCantParseCallId(sid string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongCallId, "cant parse %s call id", sid)
}

// When unique id already used in past.
func ErrNotUniqueID(uniqueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotUniqueID, "%s is not unique id, already exists", uniqueID)
}

// When call not found by unique id.
func ErrNotFoundUniqueID(uniqueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotFoundUniqueID, "%s is not found", uniqueID)
}
