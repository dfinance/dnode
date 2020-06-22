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
	EventTypeContractStatus = "contract_status"
	EventTypeMoveEvent      = "contract_events"

	// Attributes keys
	AttrKeyStatus         = "status"
	AttrKeyMajorStatus    = "major_status"
	AttrKeySubStatus      = "sub_status"
	AttrKeyMessage        = "message"
	AttrKeyType           = "type"
	AttrKeyData           = "data"
	AttrKeyGuid           = "guid"

	// Values.
	StatusDiscard = "discard"
	StatusKeep    = "keep"
	StatusError   = "error"
)

// New event with keep status.
func NewEventKeep() sdk.Event {
	return sdk.NewEvent(
		EventTypeContractStatus,
		sdk.NewAttribute(AttrKeyStatus, StatusKeep),
	)
}

// Creating discard/errors statuses.
func newEventStatus(topic string, vmStatus *vm_grpc.VMStatus) sdk.Event {
	attributes := make([]sdk.Attribute, 1)
	attributes[0] = sdk.NewAttribute(AttrKeyStatus, topic)
	if vmStatus != nil {
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(vmStatus.MajorStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(vmStatus.SubStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMessage, vmStatus.Message))
	}

	return sdk.NewEvent(
		EventTypeContractStatus,
		attributes...,
	)
}

// New event with error status.
func NewEventError(vmStatus *vm_grpc.VMStatus) sdk.Event {
	return newEventStatus(StatusError, vmStatus)
}

// New event with discard status.
func NewEventDiscard(errorStatus *vm_grpc.VMStatus) sdk.Event {
	return newEventStatus(StatusDiscard, errorStatus)
}

// Parse VM event to standard SDK event.
// In case of event data equal "struct" we don't process struct, and just keep bytes, as for any other type.
func NewEventFromVM(event *vm_grpc.VMEvent) sdk.Event {
	// TODO: implementation is wrong
	return sdk.NewEvent(
		EventTypeMoveEvent,
		//sdk.NewAttribute(AttrKeyGuid, "0x"+hex.EncodeToString(event.Key)),
		sdk.NewAttribute(AttrKeyType, VMLCSTagToStringPanic(event.EventType)),
		// we will not parse event data, as it doesn't make sense
		sdk.NewAttribute(AttrKeyData, "0x"+hex.EncodeToString(event.EventData)),
	)
}
