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
	return sdk.NewError(codeSpase, CodeErrOnlyMs, "module %q does support only multisig calls, see mshandler", moduleName)
}

// Msg.Route not found in router
func ErrRouteDoesntExist(route string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrRouteDoesntExist, "route %q not found", route)
}

// Msg.Route is empty
func ErrEmptyRoute(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrEmptyRoute, "msg route is empty for %d call", id)
}

// Msg.Type is empty (could be empty if we use MsMsg interface)
func ErrEmptyType(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrEmptyType, "msg type is empty for %d call", id)
}

// CallID not found in store / CallID > (NextCallID - 1)
func ErrWrongCallId(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongCallId, "call %d not found", id)
}

// CallID has no votes
func ErrNoVotes(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNoVotes, "no votes found for call %d", id)
}

// CallID already approved by address
func ErrCallAlreadyApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyApproved, "call %d already approved by %s", id, address)
}

// CallID not approved by address
func ErrCallNotApproved(id uint64, address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNotApproved, "call %d not approved by %s", id, address)
}

// CallID already executed
func ErrAlreadyConfirmed(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyConfirmed, "call %d already confirmed", id)
}

// TX received from not validator (POA module)
func ErrNotValidator(validator string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotValidator, "%q is not a validator", validator)
}

// CallID already rejected
func ErrAlreadyRejected(id uint64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrAlreadyRerejected, "call %d already rejected", id)
}

// UniqueID already used in past
func ErrNotUniqueID(uniqueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotUniqueID, "uniqueID %q already exists", uniqueID)
}

// Call by uniqueID not found
func ErrNotFoundUniqueID(uniqueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotFoundUniqueID, "uniqueID %q not found", uniqueID)
}
