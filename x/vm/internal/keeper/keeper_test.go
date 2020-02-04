package keeper

import "testing"

//TODO: we should move connection to vm into app, and keep connection once wb started, so then later we can test things with vm in tests.
// Test store module functional
func _TestStoreModule(t *testing.T) {
	/*input := setupTestInput(t)

	ap := vm_grpc.VMAccessPath{
		Address: input.addressBytes,
		Path:    input.pathBytes,
	}*/

	/*
		// check if store methods works
		require.NoError(t, input.vk.storeModule(input.ctx, ap, input.codeBytes))
		require.True(t, input.vk.hasModule(input.ctx, ap))

		// check double storing same module
		err := input.vk.storeModule(input.ctx, ap, input.codeBytes)
		require.Error(t, err)
		require.Equal(t, err.Codespace(), types.Codespace)
		require.Equal(t, err.Code(), sdk.CodeType(types.CodeErrModuleExists))
	*/
}
