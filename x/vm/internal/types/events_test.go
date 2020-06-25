// +build unit

package types

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

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

// Test GetSenderAddress.
func Test_GetSenderAddress(t *testing.T) {
	address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	require.EqualValues(t, common_vm.StdLibAddressShortStr, GetSenderAddress(common_vm.StdLibAddress))
	require.EqualValues(t, address.String(), GetSenderAddress(address))
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
		SenderAddress: common_vm.StdLibAddress,
		SenderModule: &vm_grpc.ModuleIdent{
			Name:    "testModule",
			Address: common_vm.Bech32ToLibra(moduleAddr),
		},
		EventType: &vm_grpc.LcsTag{
			TypeTag: vm_grpc.LcsType_LcsU64,
			StructIdent: &vm_grpc.StructIdent{
				Address: []byte{1},
				Module:  "Module_1",
				Name:    "Struct_1",
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

	sdkModuleEvent := NewEventFromVM(sdk.NewInfiniteGasMeter(), &vmEvent)
	require.Equal(t, EventTypeMoveEvent, sdkModuleEvent.Type)
	require.Len(t, sdkModuleEvent.Attributes, 4)

	// sender
	{
		attrId := 0
		require.EqualValues(t, AttrKeySenderAddress, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, GetSenderAddress(vmEvent.SenderAddress), sdkModuleEvent.Attributes[attrId].Value)
	}
	// source
	{
		attrId := 1
		require.EqualValues(t, AttrKeySource, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, GetEventSource(vmEvent.SenderModule), sdkModuleEvent.Attributes[attrId].Value)
	}
	// type
	{
		attrId := 2
		require.EqualValues(t, AttrKeyType, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), vmEvent.EventType), sdkModuleEvent.Attributes[attrId].Value)
	}
	// data
	{
		attrId := 3
		require.EqualValues(t, AttrKeyData, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, hex.EncodeToString(valBytes), sdkModuleEvent.Attributes[attrId].Value)
	}

	// Modify vmEvent: from script
	vmEvent.SenderModule = nil
	sdkScriptEvent := NewEventFromVM(sdk.NewInfiniteGasMeter(), &vmEvent)
	require.Equal(t, EventTypeMoveEvent, sdkScriptEvent.Type)
	require.Len(t, sdkScriptEvent.Attributes, 4)
	// source
	{
		attrId := 1
		require.EqualValues(t, AttrKeySource, sdkScriptEvent.Attributes[attrId].Key)
		require.EqualValues(t, SourceScript, sdkScriptEvent.Attributes[attrId].Value)
	}
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

// Processing event with out of gas.
func Test_OutOfGasProcessEvent(t *testing.T) {
	moduleAddr := make([]byte, common_vm.VMAddressLength)
	moduleAddr[common_vm.VMAddressLength-1] = 2

	value := uint64(18446744073709551615)
	valBytes := make([]byte, 8)

	// seems Move using to_le_bytes
	binary.LittleEndian.PutUint64(valBytes, value)

	vmEvent := vm_grpc.VMEvent{
		SenderAddress: common_vm.StdLibAddress,
		SenderModule: &vm_grpc.ModuleIdent{
			Name:    "testModule",
			Address: common_vm.Bech32ToLibra(moduleAddr),
		},
		EventType: &vm_grpc.LcsTag{
			TypeTag: vm_grpc.LcsType_LcsU64,
			StructIdent: &vm_grpc.StructIdent{
				Address: []byte{1},
				Module:  "Module_1",
				Name:    "Struct_1",
				TypeParams: []*vm_grpc.LcsTag{
					{
						TypeTag: vm_grpc.LcsType_LcsBool,
						StructIdent: &vm_grpc.StructIdent{
							Address:    []byte{2},
							Module:     "Module_1",
							Name:       "Struct_2",
							TypeParams: []*vm_grpc.LcsTag{
								{
									TypeTag: vm_grpc.LcsType_LcsU8,
								},
							},
						},
					},
					{
						TypeTag: vm_grpc.LcsType_LcsU128,
					},
				},
			},
		},
		EventData: valBytes,
	}

	require.PanicsWithValue(t, sdk.ErrorOutOfGas{"event type processing"}, func() {
		NewEventFromVM(sdk.NewGasMeter(1000), &vmEvent)
	})
}
