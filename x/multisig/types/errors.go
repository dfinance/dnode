// Implement error codes and messages.
package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// Msg.Route not found in router.
	ErrRouteDoesntExist = sdkErrors.Register(ModuleName, 101, "route not found")
	// CallID not found in store / CallID > (NextCallID - 1).
	ErrWrongCallId = sdkErrors.Register(ModuleName, 102, "call not found")
	// Msg.Route is empty
	ErrEmptyRoute = sdkErrors.Register(ModuleName, 103, "msg route is empty for the call")
	// Msg.Type is empty (could be empty if we use MsMsg interface).
	ErrEmptyType = sdkErrors.Register(ModuleName, 104, "msg type is empty for the call")
	// Only multisig calls supported for module.
	ErrOnlyMultisig = sdkErrors.Register(ModuleName, 105, "module does support only multisig calls (see mshandler)")

	// CallID already approved by address.
	ErrCallAlreadyApproved = sdkErrors.Register(ModuleName, 201, "call already approved")
	// CallID already executed.
	ErrAlreadyConfirmed = sdkErrors.Register(ModuleName, 202, "call already confirmed")
	// CallID already rejected.
	ErrAlreadyRejected = sdkErrors.Register(ModuleName, 203, "call already rejected")
	// CallID not approved by address.
	ErrCallNotApproved = sdkErrors.Register(ModuleName, 204, "call not approved by address")

	// CallID has no votes.
	ErrNoVotes = sdkErrors.Register(ModuleName, 301, "no votes found for the call")

	// TX received from not validator (POA module).
	ErrNotValidator = sdkErrors.Register(ModuleName, 401, "address is not a validator")
	// UniqueID already used in past.
	ErrNotUniqueID = sdkErrors.Register(ModuleName, 402, "uniqueID already exists")
	// Call by uniqueID not found
	ErrNotFoundUniqueID = sdkErrors.Register(ModuleName, 403, "uniqueID not found")
)
