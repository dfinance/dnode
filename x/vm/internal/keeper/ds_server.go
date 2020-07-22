// DataSource server implementation.
package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dfinance/dvm-proto/go/ds_grpc"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Check DSServer implements gRPC service.
var _ ds_grpc.DSServiceServer = &DSServer{}

// DSServer is a DataSource server that catches VM client data requests.
type DSServer struct {
	ds_grpc.UnimplementedDSServiceServer
	sync.Mutex
	//
	isStarted bool // check if server already listens
	//
	keeper *Keeper
	ctx    sdk.Context // current storage context
	//
	dataMiddlewares []common_vm.DSDataMiddleware // data middleware handlers
}

// GetLogger gets logger with DS server context.
func (server *DSServer) GetLogger() log.Logger {
	return server.ctx.Logger().With("module", fmt.Sprintf("x/%s/dsserver", types.ModuleName))
}

// IsStarted checks if server is already in the listen mode.
func (server *DSServer) IsStarted() bool {
	return server.isStarted
}

// RegisterDataMiddleware registers new data middleware.
func (server *DSServer) RegisterDataMiddleware(md common_vm.DSDataMiddleware) {
	server.dataMiddlewares = append(server.dataMiddlewares, md)
}

// SetContext updates server storage context.
func (server *DSServer) SetContext(ctx sdk.Context) {
	server.Lock()
	defer server.Unlock()

	server.ctx = ctx
}

// GetRaw implements gRPC service handler: returns value from the storage.
func (server *DSServer) GetRaw(_ context.Context, req *ds_grpc.DSAccessPath) (*ds_grpc.DSRawResponse, error) {
	path := &vm_grpc.VMAccessPath{
		Address: req.Address,
		Path:    req.Path,
	}

	server.GetLogger().Info(fmt.Sprintf("Get path: %s", types.StringifyVMPath(path)))

	// here go with middlewares
	blob, err := server.processMiddlewares(path)
	if err != nil {
		server.GetLogger().Error(fmt.Sprintf("Error processing middlewares for path %s: %v", types.StringifyVMPath(path), err))
		return ErrNoData(req), nil
	}

	if blob != nil {
		return &ds_grpc.DSRawResponse{Blob: blob}, nil
	}

	// we can move it to middleware later.
	if !server.keeper.hasValue(server.ctx, path) {
		server.GetLogger().Debug(fmt.Sprintf("Can't find path: %s", types.StringifyVMPath(path)))
		return ErrNoData(req), nil
	}

	server.GetLogger().Debug(fmt.Sprintf("Get path: %s", types.StringifyVMPath(path)))
	blob = server.keeper.getValue(server.ctx, path)
	server.GetLogger().Debug(fmt.Sprintf("Return values: %s\n", hex.EncodeToString(blob)))

	return &ds_grpc.DSRawResponse{Blob: blob}, nil
}

// MultiGetRaw implements gRPC service handler: returns multiple values from the storage.
func (server *DSServer) MultiGetRaw(_ context.Context, req *ds_grpc.DSAccessPaths) (*ds_grpc.DSRawResponses, error) {
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

// processMiddlewares checks that accessPath can be processed by any registered middleware.
// Contract: if {data} != nil, middleware was found.
func (server *DSServer) processMiddlewares(path *vm_grpc.VMAccessPath) (data []byte, err error) {
	for _, f := range server.dataMiddlewares {
		data, err = f(server.ctx, path)
		if err != nil || data != nil {
			return
		}
	}

	return
}

// NewDSServer creates a new DS server.
func NewDSServer(keeper *Keeper) *DSServer {
	return &DSServer{
		keeper: keeper,
	}
}

// StartServer starts DS server in the go routine.
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

// ErrNoData builds gRPC error response when data wasn't found.
func ErrNoData(path *ds_grpc.DSAccessPath) *ds_grpc.DSRawResponse {
	return &ds_grpc.DSRawResponse{
		ErrorCode:    ds_grpc.DSRawResponse_NO_DATA,
		ErrorMessage: fmt.Sprintf("data not found for access path: %s", path.String()),
	}
}
