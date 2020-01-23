package keeper

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

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

func StartServer(keeper Keeper) error {
	server := grpc.NewServer()
	vm_grpc.RegisterDSServiceServer(server, &VMServer{
		keeper: keeper,
	})

	if err := server.Serve(keeper.listener); err != nil {
		return err
	}

	return nil
}
