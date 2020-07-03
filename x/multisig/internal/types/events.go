package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// NewActiveCallsEvent returns event on multisig active calls processing start.
func NewActiveCallsEvent(blockHeight int64) sdk.Event {
	return sdk.NewEvent(
		"start-active-calls-ex",
		sdk.Attribute{
			Key:   "height",
			Value: strconv.FormatInt(blockHeight, 10),
		},
	)
}

// NewExecuteCallEvent returns event on multisig active call processing start.
func NewExecuteCallEvent(id dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		"execute-call",
		sdk.Attribute{
			Key:   "callId",
			Value: id.String(),
		},
	)
}

// NewFailedCallEvent returns event on multisig call failed.
func NewFailedCallEvent(id dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		"failed",
		sdk.Attribute{
			Key:   "callId",
			Value: id.String(),
		},
	)
}

// NewExecutedCallEvent returns event on multisig call executed.
func NewExecutedCallEvent(id dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		"executed",
		sdk.Attribute{
			Key:   "callId",
			Value: id.String(),
		},
	)
}

// NewRejectedCallsEvent returns event on multisig rejected calls processing start.
func NewRejectedCallsEvent(blockHeight int64) sdk.Event {
	return sdk.NewEvent(
		"start-rejected-calls-rem",
		sdk.Attribute{
			Key:   "callId",
			Value: strconv.FormatInt(blockHeight, 10),
		})
}

// NewRejectedCallEvent returns event on multisig call rejected.
func NewRejectedCallEvent(id dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		"reject-call",
		sdk.Attribute{
			Key:   "callId",
			Value: id.String(),
		},
	)
}
