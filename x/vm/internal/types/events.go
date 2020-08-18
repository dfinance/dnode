package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	EventTypeContractStatus = ModuleName + ".contract_status"
	EventTypeMoveEvent      = ModuleName + ".contract_events"
	//
	AttributeStatus         = "status"
	AttributeErrMajorStatus = "major_status"
	AttributeErrSubStatus   = "sub_status"
	AttributeErrMessage     = "message"
	AttributeVmEventSender  = "sender_address"
	AttributeVmEventSource  = "source"
	AttributeVmEventType    = "type"
	AttributeVmEventData    = "data"
	//
	AttributeValueStatusKeep      = "keep"
	AttributeValueStatusDiscard   = "discard"
	AttributeValueStatusError     = "error"
	AttributeValueSourceScript    = "script"
	AttributeValueSourceModuleFmt = "%s::%s"
)

// NewContractEvents creates Events on successful / failed VM execution.
// "keep" status emits two events, "discard" status emits one event.
func NewContractEvents(exec *vm_grpc.VMExecuteResponse) sdk.Events {
	if exec == nil {
		panic(fmt.Errorf("building contract sdk.Events: exec is nil"))
	}

	status := exec.GetStatus()
	if status == nil {
		panic(fmt.Errorf("building contract sdk.Events: exec.Status is nil"))
	}

	var events sdk.Events

	if status.GetError() == nil {
		events = append(events, sdk.NewEvent(
			EventTypeContractStatus,
			sdk.NewAttribute(AttributeStatus, AttributeValueStatusKeep),
		))

		return events
	}

	event := sdk.NewEvent(
		EventTypeContractStatus,
		sdk.NewAttribute(AttributeStatus, AttributeValueStatusDiscard),
	)

	if status.GetMessage() != nil {
		attributes := []sdk.Attribute{
			sdk.NewAttribute(AttributeErrMessage, status.GetMessage().GetText()),
		}

		event = event.AppendAttributes(attributes...)
	}

	events = append(events, event)

	return events
}

// NewMoveEvent converts VM event to SDK event.
// GasMeter is used to prevent long parsing (lots of nested structs).
func NewMoveEvent(gasMeter sdk.GasMeter, vmEvent *vm_grpc.VMEvent) sdk.Event {
	if vmEvent == nil {
		panic(fmt.Errorf("building Move sdk.Event: event is nil"))
	}

	// eventData: not parsed as it doesn't make sense
	return sdk.NewEvent(EventTypeMoveEvent,
		sdk.NewAttribute(AttributeVmEventSender, StringifySenderAddress(vmEvent.SenderAddress)),
		sdk.NewAttribute(AttributeVmEventSource, GetEventSourceAttribute(vmEvent.SenderModule)),
		sdk.NewAttribute(AttributeVmEventType, StringifyEventTypePanic(gasMeter, vmEvent.EventType)),
		sdk.NewAttribute(AttributeVmEventData, hex.EncodeToString(vmEvent.EventData)),
	)
}

// GetEventSourceAttribute returns SDK event attribute for VM event source (script / module) serialized to string.
func GetEventSourceAttribute(senderModule *vm_grpc.ModuleIdent) string {
	if senderModule == nil {
		return AttributeValueSourceScript
	}

	return fmt.Sprintf(AttributeValueSourceModuleFmt, StringifySenderAddress(senderModule.Address), senderModule.Name)
}
