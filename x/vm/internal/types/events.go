package types

import (
	"encoding/hex"
	"fmt"
	"strconv"

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

	statusStructAttributes := func() []sdk.Attribute {
		if exec.StatusStruct == nil {
			return nil
		}

		return []sdk.Attribute{
			sdk.NewAttribute(AttributeErrMajorStatus, strconv.FormatUint(exec.StatusStruct.MajorStatus, 10)),
			sdk.NewAttribute(AttributeErrSubStatus, strconv.FormatUint(exec.StatusStruct.SubStatus, 10)),
			sdk.NewAttribute(AttributeErrMessage, exec.StatusStruct.Message),
		}
	}

	var events sdk.Events
	switch exec.Status {
	case vm_grpc.ContractStatus_Keep:
		// "keep" event
		events = append(events, sdk.NewEvent(
			EventTypeContractStatus,
			sdk.NewAttribute(AttributeStatus, AttributeValueStatusKeep),
		))

		// "error" event
		if exec.StatusStruct != nil && exec.StatusStruct.MajorStatus != VMCodeExecuted {
			event := sdk.NewEvent(
				EventTypeContractStatus,
				sdk.NewAttribute(AttributeStatus, AttributeValueStatusError),
			)
			event = event.AppendAttributes(statusStructAttributes()...)

			events = append(events, event)
		}
	case vm_grpc.ContractStatus_Discard:
		// "discard" event
		event := sdk.NewEvent(
			EventTypeContractStatus,
			sdk.NewAttribute(AttributeStatus, AttributeValueStatusDiscard),
		)
		event = event.AppendAttributes(statusStructAttributes()...)

		events = append(events, event)
	}

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
