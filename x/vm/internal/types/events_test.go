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

// Test event build when VM return status is "keep changes".
func TestVM_KeepEvent(t *testing.T) {
	t.Parallel()

	// "keep" no error
	{
		exec := &vm_grpc.VMExecuteResponse{
			Status: vm_grpc.ContractStatus_Keep,
		}
		events := NewContractEvents(exec)

		require.Len(t, events, 1)

		event0 := events[0]
		require.Equal(t, EventTypeContractStatus, event0.Type)
		require.EqualValues(t, AttributeStatus, event0.Attributes[0].Key)
		require.EqualValues(t, AttributeValueStatusKeep, event0.Attributes[0].Value)
	}

	// "keep" with error
	{
		exec := &vm_grpc.VMExecuteResponse{
			Status: vm_grpc.ContractStatus_Keep,
			StatusStruct: &vm_grpc.VMStatus{
				MajorStatus: 100,
				SubStatus:   200,
				Message:     "this is error!!111",
			},
		}
		events := NewContractEvents(exec)

		require.Len(t, events, 2)

		event0 := events[0]
		require.Equal(t, EventTypeContractStatus, event0.Type)
		require.EqualValues(t, AttributeStatus, event0.Attributes[0].Key)
		require.EqualValues(t, AttributeValueStatusKeep, event0.Attributes[0].Value)

		event1 := events[1]
		require.Equal(t, EventTypeContractStatus, event1.Type)
		require.EqualValues(t, AttributeStatus, event1.Attributes[0].Key)
		require.EqualValues(t, AttributeValueStatusError, event1.Attributes[0].Value)
		require.EqualValues(t, AttributeErrMajorStatus, event1.Attributes[1].Key)
		require.EqualValues(t, strconv.FormatUint(100, 10), event1.Attributes[1].Value)
		require.EqualValues(t, AttributeErrSubStatus, event1.Attributes[2].Key)
		require.EqualValues(t, strconv.FormatUint(200, 10), event1.Attributes[2].Value)
		require.EqualValues(t, AttributeErrMessage, event1.Attributes[3].Key)
		require.NotEmpty(t, event1.Attributes[3].Value)
	}
}

// Test event build when VM return status is "discard changes".
func TestVM_DiscardEvent(t *testing.T) {
	t.Parallel()

	// "discard" no error
	{
		exec := &vm_grpc.VMExecuteResponse{
			Status: vm_grpc.ContractStatus_Discard,
		}
		events := NewContractEvents(exec)

		require.Len(t, events, 1)

		event0 := events[0]
		require.Equal(t, EventTypeContractStatus, event0.Type)
		require.EqualValues(t, AttributeStatus, event0.Attributes[0].Key)
		require.EqualValues(t, AttributeValueStatusDiscard, event0.Attributes[0].Value)
	}

	// "discard" with error
	{
		exec := &vm_grpc.VMExecuteResponse{
			Status: vm_grpc.ContractStatus_Discard,
			StatusStruct: &vm_grpc.VMStatus{
				MajorStatus: 0,
				SubStatus:   1,
				Message:     "this is error!!111",
			},
		}
		events := NewContractEvents(exec)

		require.Len(t, events, 1)

		event0 := events[0]
		require.Equal(t, EventTypeContractStatus, event0.Type)
		require.EqualValues(t, AttributeStatus, event0.Attributes[0].Key)
		require.EqualValues(t, AttributeValueStatusDiscard, event0.Attributes[0].Value)
		require.EqualValues(t, AttributeErrMajorStatus, event0.Attributes[1].Key)
		require.EqualValues(t, strconv.FormatUint(0, 10), event0.Attributes[1].Value)
		require.EqualValues(t, AttributeErrSubStatus, event0.Attributes[2].Key)
		require.EqualValues(t, strconv.FormatUint(1, 10), event0.Attributes[2].Value)
		require.EqualValues(t, AttributeErrMessage, event0.Attributes[3].Key)
		require.NotEmpty(t, event0.Attributes[3].Value)
	}
}

// Test StringifySenderAddress.
func TestVM_StringifySenderAddress(t *testing.T) {
	t.Parallel()

	address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	require.EqualValues(t, common_vm.StdLibAddressShortStr, StringifySenderAddress(common_vm.StdLibAddress))
	require.EqualValues(t, address.String(), StringifySenderAddress(address))
}

// Test event convertation from Move type to Cosmos.
func TestVM_NewEventFromVM(t *testing.T) {
	t.Parallel()

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

	sdkModuleEvent := NewMoveEvent(sdk.NewInfiniteGasMeter(), &vmEvent)
	require.Equal(t, EventTypeMoveEvent, sdkModuleEvent.Type)
	require.Len(t, sdkModuleEvent.Attributes, 4)

	// sender
	{
		attrId := 0
		require.EqualValues(t, AttributeVmEventSender, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, StringifySenderAddress(vmEvent.SenderAddress), sdkModuleEvent.Attributes[attrId].Value)
	}
	// source
	{
		attrId := 1
		require.EqualValues(t, AttributeVmEventSource, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, GetEventSourceAttribute(vmEvent.SenderModule), sdkModuleEvent.Attributes[attrId].Value)
	}
	// type
	{
		attrId := 2
		require.EqualValues(t, AttributeVmEventType, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), vmEvent.EventType), sdkModuleEvent.Attributes[attrId].Value)
	}
	// data
	{
		attrId := 3
		require.EqualValues(t, AttributeVmEventData, sdkModuleEvent.Attributes[attrId].Key)
		require.EqualValues(t, hex.EncodeToString(valBytes), sdkModuleEvent.Attributes[attrId].Value)
	}

	// Modify vmEvent: from script
	vmEvent.SenderModule = nil
	sdkScriptEvent := NewMoveEvent(sdk.NewInfiniteGasMeter(), &vmEvent)
	require.Equal(t, EventTypeMoveEvent, sdkScriptEvent.Type)
	require.Len(t, sdkScriptEvent.Attributes, 4)
	// source
	{
		attrId := 1
		require.EqualValues(t, AttributeVmEventSource, sdkScriptEvent.Attributes[attrId].Key)
		require.EqualValues(t, AttributeValueSourceScript, sdkScriptEvent.Attributes[attrId].Value)
	}
}

// Processing event with out of gas.
func TestVM_OutOfGasProcessEvent(t *testing.T) {
	t.Parallel()

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
							Address: []byte{2},
							Module:  "Module_1",
							Name:    "Struct_2",
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
		NewMoveEvent(sdk.NewGasMeter(1000), &vmEvent)
	})
}
