// Events types.
package types

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

const (
	// Event types.
	EventTypeDiscard = "discard"
	EventTypeKeep    = "keep"

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

// New event with discard status.
func NewEventDiscard(errorStatus *vm_grpc.VMErrorStatus) sdk.Event {
	attributes := make([]sdk.Attribute, 0)

	if errorStatus != nil {
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10)))
		attributes = append(attributes, sdk.NewAttribute(AttrKeyMessage, errorStatus.Message))
	}

	return sdk.NewEvent(
		EventTypeDiscard,
		attributes...,
	)
}

// Parse VM event to standard SDK event.
// In case of event data equal "struct" we don't process struct, and just keep bytes, as for any other type.
func NewEventFromVM(event *vm_grpc.VMEvent) sdk.Event {
	return sdk.NewEvent(
		string(event.Key),
		sdk.NewAttribute(AttrKeySequenceNumber, strconv.FormatUint(event.SequenceNumber, 10)),
		sdk.NewAttribute(AttrKeyType, VMTypeToStringPanic(event.Type.Tag)),
		// TODO: parse event data?
		sdk.NewAttribute(AttrKeyData, "0x"+hex.EncodeToString(event.EventData)),
	)
}
