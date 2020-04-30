// Implementation of Data Source (DS) server.
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

// Check type.
var _ ds_grpc.DSServiceServer = DSServer{}

// Server to catch VM data client requests.
type DSServer struct {
	ds_grpc.UnimplementedDSServiceServer

	isStarted bool // check if server already listen

	keeper *Keeper
	ctx    sdk.Context // should be careful with it, but for now we store default context

	mux sync.Mutex

	dataMiddlewares []common_vm.DSDataMiddleware
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

// Register new data middleware.
func (server *DSServer) RegisterDataMiddleware(md common_vm.DSDataMiddleware) {
	server.dataMiddlewares = append(server.dataMiddlewares, md)
}

// Process middlewares.
func (server DSServer) processMiddlewares(path *vm_grpc.VMAccessPath) (data []byte, err error) {
	for _, f := range server.dataMiddlewares {
		data, err = f(server.ctx, path)
		if err != nil || data != nil {
			return
		}
	}

	return
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

	server.Logger().Info(fmt.Sprintf("Get path: %s", types.PathToHex(path)))

	// here go with middlewares
	blob, err := server.processMiddlewares(path)
	if err != nil {
		server.Logger().Error(fmt.Sprintf("Error processing middlewares for path %s: %v", types.PathToHex(path), err))
		return ErrNoData(req), nil
	}

	if blob != nil {
		return &ds_grpc.DSRawResponse{Blob: blob}, nil
	}

	// we can move it to middleware too later.
	if !server.keeper.hasValue(server.ctx, path) {
		server.Logger().Error(fmt.Sprintf("Can't find path: %s", types.PathToHex(path)))
		return ErrNoData(req), nil
	}

	blob = server.keeper.getValue(server.ctx, path)

	server.Logger().Debug(fmt.Sprintf("Return values: %s\n", hex.EncodeToString(blob)))

	return &ds_grpc.DSRawResponse{Blob: blob}, nil
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
