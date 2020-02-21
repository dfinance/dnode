// Implementation of Data Source (DS) server.
package keeper

import (
	"context"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/ds_grpc"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"sync"
)

// Check type.
var _ ds_grpc.DSServiceServer = DSServer{}

// Server to catch VM data client requests.
type DSServer struct {
	ds_grpc.UnimplementedDSServiceServer

	isStarted bool // check if server already listen

	keeper *Keeper
	ctx    sdk.Context // should be careful with it, but for now we store default context

	mux sync.Mutex
}

// Error when no data found.
func ErrNoData(path *ds_grpc.DSAccessPath) *ds_grpc.DSRawResponse {
	return &ds_grpc.DSRawResponse{
		ErrorCode:    ds_grpc.DSRawResponse_NO_DATA,
		ErrorMessage: fmt.Sprintf("data not found for access path: %s", path.String()),
	}
}

// Server logger.
func (server *DSServer) Logger() log.Logger {
	return server.ctx.Logger().With("module", "vm")
}

// Set server context.
func (server *DSServer) SetContext(ctx sdk.Context) {
	server.mux.Lock()
	server.ctx = ctx
	server.mux.Unlock()
}

// Check if server is already in listen mode.
func (server DSServer) IsStarted() bool {
	return server.isStarted
}

// Data source processing request to return value from storage.
func (server DSServer) GetRaw(_ context.Context, req *ds_grpc.DSAccessPath) (*ds_grpc.DSRawResponse, error) {
	path := &vm_grpc.VMAccessPath{
		Address: req.Address,
		Path:    req.Path,
	}

	if !server.keeper.hasValue(server.ctx, path) {
		server.Logger().Error(fmt.Sprintf("Can't find path: %s", types.PathToHex(*path)))
		return ErrNoData(req), nil
	}

	server.Logger().Info(fmt.Sprintf("Get path: %s", types.PathToHex(*path)))

	blob := server.keeper.getValue(server.ctx, path)
	return &ds_grpc.DSRawResponse{
		Blob: blob,
	}, nil
}

// Data source processing request to return multiplay values form storage.
func (server DSServer) MultiGetRaw(_ context.Context, req *ds_grpc.DSAccessPaths) (*ds_grpc.DSRawResponses, error) {
	/*resps := &ds_grpc.DSRawResponses{
		Blobs: make([][]byte, 0),
	}

	for _, dsAccessPath := range req.Paths {
		path := &vm_grpc.VMAccessPath{
			Address: dsAccessPath.Address,
			Path:    dsAccessPath.Path,
		}

		if !server.keeper.hasValue(*server.ctx, path) {
			return nil, types.ErrDSMissedValue(*path)
		}

		blob := server.keeper.getValue(*server.ctx, path)
		resps.Blobs = append(resps.Blobs, blob)
	}

	return resps, nil*/
	return nil, status.Errorf(codes.Unimplemented, "MultiGetRaw unimplemented")
}

// Creating new DS server.
func NewDSServer(keeper *Keeper) *DSServer {
	return &DSServer{
		keeper: keeper,
	}
}

// Start DS server in go routine.
func StartServer(listener net.Listener, dsServer *DSServer) *grpc.Server {
	server := grpc.NewServer()
	ds_grpc.RegisterDSServiceServer(server, dsServer)

	go func() {
		dsServer.isStarted = true
		if err := server.Serve(listener); err != nil {
			panic(err) // should not happen during running application, after start
		}
	}()

	return server
}
