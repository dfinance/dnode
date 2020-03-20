package types

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

// Test event happens when VM return status to keep changes.
func TestNewEventKeep(t *testing.T) {
	t.Parallel()

	event := NewEventKeep()
	require.Equal(t, event.Type, EventTypeKeep)
	require.Empty(t, event.Attributes)
}

// Test event happens when VM return status discard.
func TestNewEventDiscard(t *testing.T) {
	t.Parallel()

	event := NewEventDiscard(nil)

	require.Equal(t, event.Type, EventTypeDiscard)
	require.Empty(t, event.Attributes) // has empty attributes if error status is not provided.

	errorStatus := vm_grpc.VMErrorStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	attrs := make([]sdk.Attribute, 3)
	attrs[0] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[1] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)

	event = NewEventDiscard(&errorStatus)
	require.Len(t, event.Attributes, len(attrs))

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}
}

// Test event convertation from Move type to Cosmos.
func TestNewEventFromVM(t *testing.T) {
	typeTag, err := GetVMTypeByString("U64")
	if err != nil {
		t.Fatal(err)
	}

	var value uint64 = 18446744073709551615
	valBytes := make([]byte, 8)

	// seems Move using to_le_bytes
	binary.LittleEndian.PutUint64(valBytes, value)

	vmEvent := vm_grpc.VMEvent{
		Key:            []byte("deposit"),
		SequenceNumber: 1,
		Type: &vm_grpc.VMType{
			Tag:       typeTag,
			StructTag: nil,
		},
		EventData: valBytes,
	}

	event := NewEventFromVM(&vmEvent)
	require.Equal(t, "0x"+hex.EncodeToString(vmEvent.Key), event.Type)
	require.Len(t, event.Attributes, 3)

	require.EqualValues(t, AttrKeySequenceNumber, event.Attributes[0].Key)
	require.EqualValues(t, strconv.FormatUint(vmEvent.SequenceNumber, 10), event.Attributes[0].Value)
	require.EqualValues(t, AttrKeyType, event.Attributes[1].Key)
	require.EqualValues(t, VMTypeToStringPanic(vmEvent.Type.Tag), event.Attributes[1].Value)
	require.EqualValues(t, AttrKeyData, event.Attributes[2].Key)
	require.EqualValues(t, "0x"+hex.EncodeToString(valBytes), event.Attributes[2].Value)
}

// Test event happens when VM return status with errors.
func TestNewEventError(t *testing.T) {
	errorStatus := vm_grpc.VMErrorStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	event := NewEventError(&errorStatus)

	attrs := make([]sdk.Attribute, 3)
	attrs[0] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[1] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)

	require.Len(t, event.Attributes, len(attrs))

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}

	require.EqualValues(t, event.Type, EventTypeError)
}

// Test creation event with error status.
func TestNewEventWithError(t *testing.T) {
	event := newEventErrorStatus(EventTypeError, nil)

	require.Empty(t, event.Attributes)
	require.Equal(t, EventTypeError, event.Type)

	errorStatus := vm_grpc.VMErrorStatus{
		MajorStatus: 0,
		SubStatus:   0,
		Message:     "this is error!!111",
	}

	event = newEventErrorStatus(EventTypeDiscard, &errorStatus)

	require.Equal(t, EventTypeDiscard, event.Type)

	attrs := make([]sdk.Attribute, 3)
	attrs[0] = sdk.NewAttribute(AttrKeyMajorStatus, strconv.FormatUint(errorStatus.MajorStatus, 10))
	attrs[1] = sdk.NewAttribute(AttrKeySubStatus, strconv.FormatUint(errorStatus.SubStatus, 10))
	attrs[2] = sdk.NewAttribute(AttrKeyMessage, errorStatus.Message)
	require.Len(t, event.Attributes, len(attrs))

	for i, attr := range attrs {
		require.EqualValuesf(t, []byte(attr.Key), event.Attributes[i].Key, "incorrect attribute key for event discard at position %d", i)
		require.EqualValuesf(t, []byte(attr.Value), event.Attributes[i].Value, "incorrect attribute key for event discard at position %d", i)
	}
}
