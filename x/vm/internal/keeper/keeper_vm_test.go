package keeper

import (
	"bytes"
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// TODO: change listener logic to don't close it here?

// Check storage set value functional.
func TestSetValue(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	ap := &vm_grpc.VMAccessPath{
		Address: input.addressBytes,
		Path:    input.pathBytes,
	}

	input.vk.setValue(input.ctx, ap, input.valueBytes)
	value := input.vk.getValue(input.ctx, ap)

	require.True(t, bytes.Equal(input.valueBytes, value))

	isExists := input.vk.hasValue(input.ctx, ap)
	require.True(t, isExists)
}

// Check get value from storage functional.
func TestGetValue(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	ap := randomPath()
	input.vk.setValue(input.ctx, ap, input.valueBytes)

	value := input.vk.getValue(input.ctx, ap)
	require.Equal(t, input.valueBytes, value)

	notExistsPath := randomPath()

	var nilBytes []byte
	value = input.vk.getValue(input.ctx, notExistsPath)
	require.EqualValues(t, nilBytes, value)
}

// Check has value functional.
func TestHasValue(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	ap := randomPath()

	input.vk.setValue(input.ctx, ap, input.valueBytes)

	isExists := input.vk.hasValue(input.ctx, ap)
	require.True(t, isExists)

	wrongAp := randomPath()
	isExists = input.vk.hasValue(input.ctx, wrongAp)
	require.False(t, isExists)
}

// Check deletion of key in storage.
func TestDelValue(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	var emptyBytes []byte

	ap := randomPath()
	input.vk.setValue(input.ctx, ap, input.valueBytes)

	value := input.vk.getValue(input.ctx, ap)
	require.EqualValues(t, input.valueBytes, value)

	isExists := input.vk.hasValue(input.ctx, ap)
	require.True(t, isExists)

	input.vk.delValue(input.ctx, ap)

	value = input.vk.getValue(input.ctx, ap)
	require.EqualValues(t, emptyBytes, value)

	isExists = input.vk.hasValue(input.ctx, ap)
	require.False(t, isExists)
}

// Check process execution (response from VM) functional.
func TestProcessExecution(t *testing.T) {
	// ignoring gas for now.
	input := setupTestInput(true)
	defer closeInput(input)

	resp := &vm_grpc.VMExecuteResponse{
		Status: vm_grpc.ContractStatus_Discard,
		StatusStruct: &vm_grpc.VMErrorStatus{
			MajorStatus: 1,
			SubStatus:   250,
			Message:     "this is another errorr!!!1111",
		},
	}

	input.vk.processExecution(input.ctx, resp)

	events := input.ctx.EventManager().Events()
	event := types.NewEventDiscard(resp.StatusStruct)

	require.Len(t, events, 1)
	require.Equal(t, event.Type, events[0].Type)
	require.Equal(t, event.Attributes, events[0].Attributes)

	// discard without status
	resp = &vm_grpc.VMExecuteResponse{
		Status: vm_grpc.ContractStatus_Discard,
	}

	ctx := input.ctx.WithEventManager(sdk.NewEventManager())
	input.vk.processExecution(ctx, resp)

	events = ctx.EventManager().Events()
	event = types.NewEventDiscard(nil)

	require.Len(t, events, 1)
	require.Nil(t, events[0].Attributes)

	require.Equal(t, event, events[0])

	// status keep
	resp = &vm_grpc.VMExecuteResponse{
		Status: vm_grpc.ContractStatus_Keep,
	}

	ctx = input.ctx.WithEventManager(sdk.NewEventManager())
	input.vk.processExecution(ctx, resp)

	events = ctx.EventManager().Events()
	event = types.NewEventKeep()

	require.Len(t, events, 1)
	require.Equal(t, event, events[0])

	// write set & events
	var u64Value uint64 = 100
	u64Bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(u64Bytes, u64Value)

	respEvents := make([]*vm_grpc.VMEvent, 2)
	respEvents[0] = &vm_grpc.VMEvent{
		Key:            []byte("test 1"),
		SequenceNumber: 0,
		Type: &vm_grpc.VMType{
			Tag: vm_grpc.VMTypeTag_ByteArray,
		},
		EventData: randomValue(32),
	}
	respEvents[1] = &vm_grpc.VMEvent{
		Key:            []byte("test 2"),
		SequenceNumber: 1,
		Type: &vm_grpc.VMType{
			Tag: vm_grpc.VMTypeTag_U64,
		},
		EventData: u64Bytes,
	}

	wbEvents := make(sdk.Events, 2)
	wbEvents[0] = types.NewEventFromVM(respEvents[0])
	wbEvents[1] = types.NewEventFromVM(respEvents[1])

	writeSet := make([]*vm_grpc.VMValue, 2)
	writeSet[0] = &vm_grpc.VMValue{
		Type:  vm_grpc.VmWriteOp_Value,
		Value: randomValue(32),
		Path:  randomPath(),
	}
	writeSet[1] = &vm_grpc.VMValue{
		Type:  vm_grpc.VmWriteOp_Value,
		Value: randomValue(16),
		Path:  randomPath(),
	}

	resp = &vm_grpc.VMExecuteResponse{
		WriteSet: writeSet,
		Events:   respEvents,
		Status:   vm_grpc.ContractStatus_Keep,
	}

	ctx = input.ctx.WithEventManager(sdk.NewEventManager())
	input.vk.processExecution(ctx, resp)
	events = ctx.EventManager().Events()

	// check that everything fine with write set
	for _, write := range writeSet {
		require.True(t, input.vk.hasValue(input.ctx, write.Path))
		require.Equal(t, write.Value, input.vk.getValue(input.ctx, write.Path))
	}

	require.Len(t, events, len(wbEvents)+1)

	for i, event := range events[1:] {
		require.Equal(t, wbEvents[i].Type, event.Type)

		for j, attr := range event.Attributes {
			require.EqualValues(t, wbEvents[i].Attributes[j].Key, attr.Key)
			require.EqualValues(t, wbEvents[i].Attributes[j].Value, attr.Value)
		}
	}

	// check deletion
	writeSet[1] = &vm_grpc.VMValue{
		Type: vm_grpc.VmWriteOp_Deletion,
		Path: writeSet[1].Path,
	}

	resp = &vm_grpc.VMExecuteResponse{
		WriteSet: writeSet,
		Status:   vm_grpc.ContractStatus_Keep,
	}

	ctx = input.ctx.WithEventManager(sdk.NewEventManager())
	input.vk.processExecution(ctx, resp)
	events = ctx.EventManager().Events()

	require.Len(t, events, 1)

	require.False(t, input.vk.hasValue(input.ctx, writeSet[1].Path))
	require.Nil(t, input.vk.getValue(input.ctx, writeSet[1].Path))
}

// Check returned write set procession.
func TestProcessWriteSet(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	writeSet := make([]*vm_grpc.VMValue, 2)
	writeSet[0] = &vm_grpc.VMValue{
		Type:  vm_grpc.VmWriteOp_Value,
		Value: randomValue(32),
		Path:  randomPath(),
	}
	writeSet[1] = &vm_grpc.VMValue{
		Type:  vm_grpc.VmWriteOp_Value,
		Value: randomValue(16),
		Path:  randomPath(),
	}

	input.vk.processWriteSet(input.ctx, writeSet)

	// now read storage and check results
	values := make([][]byte, 2)
	values[0] = input.vk.getValue(input.ctx, writeSet[0].Path)
	values[1] = input.vk.getValue(input.ctx, writeSet[1].Path)

	for i, write := range writeSet {
		require.True(t, input.vk.hasValue(input.ctx, write.Path))
		require.Equal(t, write.Value, values[i])
	}

	// check delete op
	delSet := make([]*vm_grpc.VMValue, 2)
	delSet[0] = &vm_grpc.VMValue{
		Type: vm_grpc.VmWriteOp_Deletion,
		Path: writeSet[0].Path,
	}
	delSet[1] = &vm_grpc.VMValue{
		Type: vm_grpc.VmWriteOp_Deletion,
		Path: writeSet[1].Path,
	}

	input.vk.processWriteSet(input.ctx, delSet)

	for _, del := range delSet {
		require.False(t, input.vk.hasValue(input.ctx, del.Path))
		value := input.vk.getValue(input.ctx, del.Path)
		require.Nil(t, value)
	}
}

// Status keep (4001) still doesn't contains error.
func TestExecStatusKeeperNotAnError(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	errorStatus := vm_grpc.VMErrorStatus{
		MajorStatus: types.VMCodeExecuted,
		SubStatus:   0,
		Message:     "",
	}

	resp := &vm_grpc.VMExecuteResponse{
		WriteSet:     nil,
		Events:       nil,
		Status:       vm_grpc.ContractStatus_Keep,
		StatusStruct: &errorStatus,
	}

	input.vk.processExecution(input.ctx, resp)
	events := input.ctx.EventManager().Events()

	require.EqualValues(t, types.EventTypeKeep, events[0].Type)

	for _, event := range events {
		require.NotEqual(t, types.EventTypeError, event.Type)
	}
}

// When status still keep but returns error (and it could exists).
func TestExecKeepAndError(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	errorStatus := vm_grpc.VMErrorStatus{
		MajorStatus: 16,
		SubStatus:   0,
		Message:     "aborted error!11111!1!!!",
	}

	resp := &vm_grpc.VMExecuteResponse{
		WriteSet:     nil,
		Events:       nil,
		Status:       vm_grpc.ContractStatus_Keep,
		StatusStruct: &errorStatus,
	}

	input.vk.processExecution(input.ctx, resp)
	events := input.ctx.EventManager().Events()

	require.EqualValues(t, types.EventTypeKeep, events[0].Type)

	errEvent := events[1]
	require.Equal(t, types.EventTypeError, errEvent.Type)

	shouldBeErr := types.NewEventError(&errorStatus)
	require.True(t, reflect.DeepEqual(shouldBeErr, errEvent))
}
