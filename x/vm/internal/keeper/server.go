package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc"
	"net"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/ds_grpc"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// check type.
var _ ds_grpc.DSServiceServer = &DSServer{}

// Server to catch VM data client requests.
type DSServer struct {
	ds_grpc.UnimplementedDSServiceServer

	keeper *Keeper
	ctx    *sdk.Context // should be careful with it, but for now we store default context
}

// As we expect before call that VM server can ask for data, we should call SetContext before every request to
// VM.
// Later we should check before every request that context is setup and normal.
func (server *DSServer) SetContext(ctx *sdk.Context) {
	server.ctx = ctx
}

func (server DSServer) GetRaw(_ context.Context, req *ds_grpc.DSAccessPath) (*ds_grpc.DSRawResponse, error) {
	path := &vm_grpc.VMAccessPath{
		Address: req.Address,
		Path:    req.Path,
	}

	if !server.keeper.hasValue(*server.ctx, path) {
		return nil, types.ErrDSMissedValue(*path)
	}

	blob := server.keeper.getValue(*server.ctx, path)
	return &ds_grpc.DSRawResponse{
		Blob: blob,
	}, nil
}

func (server DSServer) MultiGetRaw(_ context.Context, req *ds_grpc.DSAccessPaths) (*ds_grpc.DSRawResponses, error) {
	resps := &ds_grpc.DSRawResponses{
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

	return resps, nil
}

func NewDSServer(keeper *Keeper) *DSServer {
	return &DSServer{
		keeper: keeper,
	}
}

func StartServer(listener net.Listener, dsServer *DSServer) {
	server := grpc.NewServer()
	ds_grpc.RegisterDSServiceServer(server, dsServer)

	if err := server.Serve(listener); err != nil {
		panic(err) // should not happen during running application, after start
	}
}
