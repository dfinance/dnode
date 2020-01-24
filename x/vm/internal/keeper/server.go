package keeper

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

// check type.
var _ vm_grpc.DSServiceServer = &VMServer{}

// Server to catch VM data client requests.
type VMServer struct {
	vm_grpc.UnimplementedDSServiceServer

	keeper Keeper
}

func (*VMServer) GetRaw(ctx context.Context, req *vm_grpc.DSAccessPath) (*vm_grpc.DSRawResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRaw not implemented")
}

func (*VMServer) MultiGetRaw(ctx context.Context, req *vm_grpc.DSAccessPaths) (*vm_grpc.DSRawResponses, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MultiGetRaw not implemented")
}

func StartServer(keeper Keeper) {
	server := grpc.NewServer()
	vm_grpc.RegisterDSServiceServer(server, &VMServer{
		keeper: keeper,
	})

	if err := server.Serve(keeper.listener); err != nil {
		panic(err) // should not happen during running application, after start
	}
}
