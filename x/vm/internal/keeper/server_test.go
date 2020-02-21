package keeper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/ds_grpc"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
)

// Initialize connection to DS server.
func getClient(t *testing.T, listener *bufconn.Listener) ds_grpc.DSServiceClient {
	dsConn, err := grpc.DialContext(context.TODO(), "", grpc.WithContextDialer(GetBufDialer(listener)), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	return ds_grpc.NewDSServiceClient(dsConn)
}

// Test set context for server.
func TestDSServer_SetContext(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	input.vk.dsServer.SetContext(input.ctx)
	require.EqualValues(t, input.ctx, input.vk.dsServer.ctx)
}

// Test get raw data from server.
func TestDSServer_GetRaw(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	rawServer := StartServer(input.vk.listener, input.vk.dsServer)
	defer rawServer.Stop()

	input.vk.dsServer.SetContext(input.ctx)

	client := getClient(t, input.dsListener)

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
	input := setupTestInput(true)
	defer closeInput(input)

	rawServer := StartServer(input.vk.listener, input.vk.dsServer)
	defer rawServer.Stop()

	input.vk.dsServer.SetContext(input.ctx)

	client := getClient(t, input.dsListener)
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
	require.Nil(t, resp)
	require.Error(t, err)

	/*for i, val := range resp.Blobs {
		require.EqualValues(t, values[i], val)
	}*/
}
