// +build integ

package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/helpers/tests/mockdvm"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

func TestVMKeeper_RetryMechanism(t *testing.T) {
	t.Parallel()

	input := newTestInput(true)
	defer input.Stop()
	ctx, keeper := input.ctx, input.vk

	// start mockDVM gRPC server (module publisher)
	listenerAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err, "geting free TCP port for MockDVM listener")

	mockDvmListener, err := helpers.GetGRpcNetListener(listenerAddr)
	require.NoError(t, err, "creating MockDVM listener")

	mockDvmServer := mockdvm.StartMockDVMService(mockDvmListener)
	defer mockDvmServer.Stop()

	// create mockDVM gRPC client and rewrite test keeper's one
	mockDvmCLient, err := helpers.GetGRpcClientConnection(listenerAddr, 1*time.Second)
	require.NoError(t, err, "creating MockDVM client")
	keeper.rawClient = mockDvmCLient
	keeper.client = NewVMClient(mockDvmCLient)

	deployReq, err := NewDeployRequest(ctx, types.MsgDeployModule{
		Signer: common_vm.StdLibAddress,
		Module: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	})
	require.NoError(t, err, "creating deployRequest")

	// ok: in one attempt (infinite settings)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 0, 0

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetExecutionDelay(50 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// ok: in one attempt (settings with limit)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 1, 5000

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetExecutionDelay(10 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// ok: in one attempt (without request timeout)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 1, 0

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetExecutionDelay(500 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// ok: in multiple attempts (with request timeout)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 10, 5000

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetSequentialFailingCount(5)
		mockDvmServer.SetExecutionDelay(10 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// ok: in multiple attempts (without request timeout)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 10, 0

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetSequentialFailingCount(5)
		mockDvmServer.SetExecutionDelay(100 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// ok: in one attempt with long response (without limits)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 0, 0

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetExecutionDelay(3000 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.NoError(t, err)
	}

	// fail: by timeout (deadline)
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 5, 100

		mockDvmServer.SetExecutionOK()
		mockDvmServer.SetResponseOK()
		mockDvmServer.SetExecutionDelay(200 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	}

	// fail: by attempts
	{
		keeper.config.MaxAttempts, keeper.config.ReqTimeoutInMs = 5, 0

		mockDvmServer.SetExecutionFail()
		mockDvmServer.SetExecutionDelay(50 * time.Millisecond)

		_, err := keeper.sendExecuteReq(ctx, deployReq, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failing gRPC execution")
	}
}
