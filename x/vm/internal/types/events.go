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
	AttributeStatus             = "status"
	AttributeErrMajorStatus     = "major_status"
	AttributeErrSubStatus       = "sub_status"
	AttributeErrMessage         = "message"
	AttributeErrLocationAddress = "location_address"
	AttributeErrLocationModule  = "location_module"
	AttributeVmEventSender      = "sender_address"
	AttributeVmEventSource      = "source"
	AttributeVmEventType        = "type"
	AttributeVmEventData        = "data"
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

	if status.GetError() == nil {
		return sdk.Events{
			sdk.NewEvent(
				EventTypeContractStatus,
				sdk.NewAttribute(AttributeStatus, AttributeValueStatusKeep),
				sdk.NewAttribute(AttributeErrMajorStatus, VMExecutedCode),
				sdk.NewAttribute(AttributeErrSubStatus, "0"),
			),
		}
	}

	// Allocate memory for 5 possible attributes: status, abort location 2 attributes, major and sub codes
	attributes := make([]sdk.Attribute, 1, 5)
	attributes[0] = sdk.NewAttribute(AttributeStatus, AttributeValueStatusDiscard)

	if sErr := status.GetError(); sErr != nil {
		var majorStatus uint64
		subStatus := uint64(0)

		switch sErr.(type) {
		case *vm_grpc.VMStatus_Abort:
			majorStatus = VMAbortedCode
			abort := sErr.(*vm_grpc.VMStatus_Abort).Abort
			if abort != nil {
				subStatus = abort.GetAbortCode()
				attributes = append(attributes, processAbortLocation(abort.AbortLocation)...)
			}
		case *vm_grpc.VMStatus_ExecutionFailure:
			executionFailure := sErr.(*vm_grpc.VMStatus_ExecutionFailure).ExecutionFailure
			if executionFailure != nil {
				majorStatus = executionFailure.GetStatusCode()
				attributes = append(attributes, processAbortLocation(executionFailure.AbortLocation)...)
			}
		case *vm_grpc.VMStatus_MoveError:
			majorStatus = sErr.(*vm_grpc.VMStatus_MoveError).MoveError.GetStatusCode()
		}

		attributes = append(
			attributes,
			sdk.NewAttribute(AttributeErrMajorStatus, strconv.FormatUint(majorStatus, 10)),
			sdk.NewAttribute(AttributeErrSubStatus, strconv.FormatUint(subStatus, 10)),
		)

		if status.GetMessage() != nil {
			attributes = append(attributes, sdk.NewAttribute(AttributeErrMessage, status.GetMessage().GetText()))
		}
	}

	return sdk.Events{sdk.NewEvent(EventTypeContractStatus, attributes...)}
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

func processAbortLocation(abortLocation *vm_grpc.AbortLocation) (attributes []sdk.Attribute) {
	if abortLocation != nil {
		attributes = make([]sdk.Attribute, 0, 2)
		if abortLocation.GetAddress() != nil {
			address := abortLocation.GetAddress()
			attributes = append(attributes, sdk.NewAttribute(AttributeErrLocationAddress, string(address)))
		}

		if abortLocation.GetModule() != "" {
			attributes = append(attributes, sdk.NewAttribute(AttributeErrLocationModule, abortLocation.GetModule()))
		}
	}

	return
}
