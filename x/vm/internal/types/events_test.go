// +build unit

package types

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
)

// Test event happens when VM return status to keep changes.
func TestNewEventKeep(t *testing.T) {
	t.Parallel()

	event := NewEventKeep()
	require.Equal(t, EventTypeContractStatus, event.Type)
	require.EqualValues(t, AttrKeyStatus, event.Attributes[0].Key)
	require.EqualValues(t, StatusKeep, event.Attributes[0].Value)
}

// Test event happens when VM return status discard.
func TestNewEventDiscard(t *testing.T) {
	t.Parallel()

	event := NewEventDiscard(nil)

	require.Equal(t, EventTypeContractStatus, event.Type)
	require.EqualValues(t, AttrKeyStatus, event.Attributes[0].Key)
	require.EqualValues(t, StatusDiscard, event.Attributes[0].Value)

	errorStatus := vm_grpc.VMStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	attrs := make([]sdk.Attribute, 4)
	attrs[0] = sdk.NewAttribute(AttrKeyStatus, StatusDiscard)
	attrs[1] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[3] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)

	event = NewEventDiscard(&errorStatus)
	require.Len(t, event.Attributes, len(attrs))
	require.Equal(t, EventTypeContractStatus, event.Type)

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}
}

// Test event convertation from Move type to Cosmos.
func TestNewEventFromVM(t *testing.T) {
	moduleAddr := make([]byte, common_vm.VMAddressLength)
	moduleAddr[common_vm.VMAddressLength-1] = 2

	value := uint64(18446744073709551615)
	valBytes := make([]byte, 8)

	// seems Move using to_le_bytes
	binary.LittleEndian.PutUint64(valBytes, value)

	vmEvent := vm_grpc.VMEvent{
		SenderAddress: common_vm.Bech32ToLibra(common_vm.StdLibAddress),
		SenderModule: &vm_grpc.ModuleIdent{
			Name:    "testModule",
			Address: common_vm.Bech32ToLibra(moduleAddr),
		},
		EventType: &vm_grpc.LcsTag{
			TypeTag: vm_grpc.LcsType_LcsU64,
			StructIdent: &vm_grpc.StructIdent{
				Address:    []byte{1},
				Module:     "Module_1",
				Name:       "Struct_1",
				TypeParams: []*vm_grpc.LcsTag{
					{
						TypeTag: vm_grpc.LcsType_LcsBool,
					},
					{
						TypeTag: vm_grpc.LcsType_LcsU128,
					},
				},
			},
		},
		EventData: valBytes,
	}

	sdkEvent := NewEventFromVM(&vmEvent)
	require.Equal(t, EventTypeMoveEvent, sdkEvent.Type)
	require.Len(t, sdkEvent.Attributes, 5)

	// sender
	require.EqualValues(t, "0x"+hex.EncodeToString(vmEvent.SenderAddress), sdkEvent.Attributes[0].Value)
	// type
	// TODO: sdkEvent.Attributes[1] is omitted
	// data
	require.EqualValues(t, AttrKeyData, sdkEvent.Attributes[2].Key)
	require.EqualValues(t, "0x"+hex.EncodeToString(valBytes), sdkEvent.Attributes[2].Value)
	// module
	require.EqualValues(t, AttrKeyModuleName, sdkEvent.Attributes[3].Key)
	require.EqualValues(t, vmEvent.SenderModule.Name, sdkEvent.Attributes[3].Value)
	require.EqualValues(t, AttrKeyModuleAddress, sdkEvent.Attributes[4].Key)
	require.EqualValues(t, "0x"+hex.EncodeToString(vmEvent.SenderModule.Address), sdkEvent.Attributes[4].Value)
}

// Test event happens when VM return status with errors.
func TestNewEventError(t *testing.T) {
	errorStatus := vm_grpc.VMStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	event := NewEventError(&errorStatus)

	attrs := make([]sdk.Attribute, 4)
	attrs[0] = sdk.NewAttribute(AttrKeyStatus, StatusError)
	attrs[1] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[3] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)

	require.Len(t, event.Attributes, len(attrs))

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}

	require.EqualValues(t, EventTypeContractStatus, event.Type)
}

// Test creation event with error status.
func TestNewEventWithError(t *testing.T) {
	event := newEventStatus(StatusKeep, nil)

	require.Equal(t, EventTypeContractStatus, event.Type)
	require.EqualValues(t, AttrKeyStatus, event.Attributes[0].Key)
	require.EqualValues(t, StatusKeep, event.Attributes[0].Value)

	errorStatus := vm_grpc.VMStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	event = newEventStatus(StatusDiscard, &errorStatus)
	require.Equal(t, EventTypeContractStatus, event.Type)

	attrs := make([]sdk.Attribute, 4)
	attrs[0] = sdk.NewAttribute(AttrKeyStatus, StatusDiscard)
	attrs[1] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[3] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)
	require.Len(t, event.Attributes, len(attrs))

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}
}
