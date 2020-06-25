// Events types.
package types

import (
	"bytes"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
)

const (
	// Event types.
	EventTypeContractStatus = "contract_status"
	EventTypeMoveEvent      = "contract_events"

	// Attributes keys.
	AttrKeyStatus        = "status"
	AttrKeyMajorStatus   = "major_status"
	AttrKeySubStatus     = "sub_status"
	AttrKeyMessage       = "message"
	AttrKeyType          = "type"
	AttrKeySenderAddress = "sender_address"
	AttrKeyModuleName    = "module_name"
	AttrKeyModuleAddress = "module_address"
	AttrKeyData          = "data"

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

// Get sender address 0x1 or wallet1...
func GetSenderAddress(addr []byte) string {
	if bytes.Equal(addr, common_vm.StdLibAddress) {
		return common_vm.StdLibAddressShortStr
	} else {
		return sdk.AccAddress(addr).String()
	}
}

// Parse VM event to standard SDK event.
// In case of event data equal "struct" we don't process struct, and just keep bytes, as for any other type.
func NewEventFromVM(gasMeter sdk.GasMeter, event *vm_grpc.VMEvent) sdk.Event {
	// eventData: not parsed as it doesn't make sense
	attrs := []sdk.Attribute{
		sdk.NewAttribute(AttrKeySenderAddress, GetSenderAddress(event.SenderAddress)),
		sdk.NewAttribute(AttrKeyType, StringifyEventTypePanic(gasMeter, event.EventType)),
		sdk.NewAttribute(AttrKeyData, hex.EncodeToString(event.EventData)),
	}

	if event.SenderModule != nil {
		attrs = append(attrs, sdk.NewAttribute(AttrKeyModuleName, event.SenderModule.Name))
		attrs = append(attrs, sdk.NewAttribute(AttrKeyModuleAddress, GetSenderAddress(event.SenderModule.Address)))
	}

	return sdk.NewEvent(EventTypeMoveEvent, attrs...)
}
