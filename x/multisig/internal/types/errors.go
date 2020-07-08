package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// Call errors
	ErrWrongCallId       = sdkErrors.Register(ModuleName, 200, "invalid call: ID")
	ErrWrongCallUniqueId = sdkErrors.Register(ModuleName, 201, "invalid call: uniqueID")
	// Multisig message specific error
	ErrWrongMsg      = sdkErrors.Register(ModuleName, 300, "invalid multisig message")
	ErrWrongMsgRoute = sdkErrors.Register(ModuleName, 301, "invalid multisig message: route")
	ErrWrongMsgType  = sdkErrors.Register(ModuleName, 302, "invalid multisig message: type")
	// Vote errors
	ErrVoteAlreadyApproved  = sdkErrors.Register(ModuleName, 400, "call already approved")
	ErrVoteAlreadyConfirmed = sdkErrors.Register(ModuleName, 401, "call already confirmed")
	ErrVoteAlreadyRejected  = sdkErrors.Register(ModuleName, 402, "call already rejected")
	ErrVoteNoVotes          = sdkErrors.Register(ModuleName, 403, "no votes found for the call")
	ErrVoteNotApproved      = sdkErrors.Register(ModuleName, 404, "call not approved by address")
	// POA
	ErrPoaNotValidator = sdkErrors.Register(ModuleName, 500, "address is not a validator")
)
