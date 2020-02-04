package keeper

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"wings-blockchain/cmd/config"
	"wings-blockchain/x/vm/internal/types/ds_grpc"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// Initialize connection to DS server.
func getClient(t *testing.T) ds_grpc.DSServiceClient {
	config := config.DefaultVMConfig()

	dsConn, err := grpc.Dial(config.DataListen, grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	return ds_grpc.NewDSServiceClient(dsConn)
}

// Test set context for server.
func TestDSServer_SetContext(t *testing.T) {
	input := setupTestInput()
	defer input.vk.listener.Close()

	require.Nil(t, input.vk.dsServer.ctx)

	input.vk.dsServer.SetContext(&input.ctx)
	require.EqualValues(t, input.ctx, *input.vk.dsServer.ctx)
}

// Test get raw data from server.
func TestDSServer_GetRaw(t *testing.T) {
	input := setupTestInput()
	rawServer := StartServer(input.vk.listener, input.vk.dsServer)
	defer rawServer.Stop()

	input.vk.dsServer.SetContext(&input.ctx)

	client := getClient(t)

	value := randomValue(32)
	ap := randomPath()

	input.vk.setValue(input.ctx, ap, value)

	connCtx := context.Background()

	resp, err := client.GetRaw(connCtx, &ds_grpc.DSAccessPath{
		Address: ap.Address,
		Path:    ap.Path,
	})
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, value, resp.Blob)
}

// Test get multiraw data from server.
func TestDSServer_MultiGetRaw(t *testing.T) {
	input := setupTestInput()
	rawServer := StartServer(input.vk.listener, input.vk.dsServer)
	defer rawServer.Stop()

	input.vk.dsServer.SetContext(&input.ctx)

	client := getClient(t)
	argsCount := 3
	req := &ds_grpc.DSAccessPaths{
		Paths: make([]*ds_grpc.DSAccessPath, argsCount),
	}
	values := make([][]byte, argsCount)

	for i := 0; i < len(req.Paths); i++ {
		path := &vm_grpc.VMAccessPath{
			Address: randomValue(32),
			Path:    randomValue(32),
		}

		values[i] = randomValue(8 * (i + 1))
		req.Paths[i] = &ds_grpc.DSAccessPath{
			Address: path.Address,
			Path:    path.Path,
		}

		input.vk.setValue(input.ctx, path, values[i])
	}

	connCtx := context.Background()
	resp, err := client.MultiGetRaw(connCtx, req)
	if err != nil {
		t.Fatal(err)
	}

	for i, val := range resp.Blobs {
		require.EqualValues(t, values[i], val)
	}
}
