package keeper

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// TODO: change listener logic to don't close it here?

// Check storage set value functional.
func TestSetValue(t *testing.T) {
	input := setupTestInput()
	defer input.vk.listener.Close()

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
	input := setupTestInput()
	defer input.vk.listener.Close()

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
	input := setupTestInput()
	defer input.vk.listener.Close()

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
	input := setupTestInput()
	defer input.vk.listener.Close()

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
