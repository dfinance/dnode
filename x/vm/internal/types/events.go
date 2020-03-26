// Events types.
package types

import (
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	// Event types.
	EventTypeDiscard = "discard"
	EventTypeKeep    = "keep"
	EventTypeError   = "error"

	// Attributes keys
	AttrKeyMajorStatus    = "major_status"
	AttrKeySubStatus      = "sub_status"
	AttrKeyMessage        = "message"
	AttrKeySequenceNumber = "sequence_number"
	AttrKeyType           = "type"
	AttrKeyData           = "data"
)

// New event with keep status.
func NewEventKeep() sdk.Event {
	return sdk.NewEvent(
		EventTypeKeep,
	)
}

// Creating discard/errors statuses.
func newEventStatus(topic string, vmStatus *vm_grpc.VMStatus) sdk.Event {
	attributes := make([]sdk.Attribute, 0)
	if vmStatus != nil {
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(vmStatus.MajorStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(vmStatus.SubStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMessage, vmStatus.Message))
	}

	return sdk.NewEvent(
		topic,
		attributes...,
	)
}

// New event with error status.
func NewEventError(vmStatus *vm_grpc.VMStatus) sdk.Event {
	return newEventStatus(EventTypeError, vmStatus)
}

// New event with discard status.
func NewEventDiscard(errorStatus *vm_grpc.VMStatus) sdk.Event {
	return newEventStatus(EventTypeDiscard, errorStatus)
}

// Parse VM event to standard SDK event.
// In case of event data equal "struct" we don't process struct, and just keep bytes, as for any other type.
func NewEventFromVM(event *vm_grpc.VMEvent) sdk.Event {
	return sdk.NewEvent(
		"0x"+hex.EncodeToString(event.Key),
		sdk.NewAttribute(AttrKeySequenceNumber, strconv.FormatUint(event.SequenceNumber, 10)),
		sdk.NewAttribute(AttrKeyType, VMTypeToStringPanic(event.Type.Tag)),
		// we will not parse event data, as it doesn't make sense
		sdk.NewAttribute(AttrKeyData, "0x"+hex.EncodeToString(event.EventData)),
	)
}
