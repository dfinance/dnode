package mockdvm

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
)

type MockDVM struct {
	sync.Mutex
	server        *grpc.Server
	failExecution bool
	failResponse  bool
	failCountdown uint
	execDelay     time.Duration
}

func (s *MockDVM) SetExecutionFail() { s.failExecution = true }
func (s *MockDVM) SetExecutionOK()   { s.failExecution = false }
func (s *MockDVM) SetResponseFail()  { s.failResponse = true }
func (s *MockDVM) SetResponseOK()    { s.failResponse = false }
func (s *MockDVM) SetExecutionDelay(dur time.Duration) {
	s.execDelay = dur
}
func (s *MockDVM) SetSequentialFailingCount(cnt uint) {
	s.failCountdown = cnt
}
func (s *MockDVM) Stop() {
	if s.server != nil {
		s.server.Stop()
	}
}

func (s *MockDVM) PublishModule(ctx context.Context, in *vm_grpc.VMPublishModule) (*vm_grpc.VMExecuteResponse, error) {
	s.Lock()
	defer s.Unlock()

	time.Sleep(s.execDelay)

	if s.failExecution || s.failCountdown > 0 {
		if s.failCountdown > 0 {
			s.failCountdown--
		}

		return nil, grpcStatus.Errorf(codes.Internal, "failing gRPC execution")
	}

	resp := &vm_grpc.VMExecuteResponse{}
	if !s.failResponse {
		resp = &vm_grpc.VMExecuteResponse{
			WriteSet:     nil,
			Events:       nil,
			GasUsed:      1,
			Status:       vm_grpc.ContractStatus_Discard,
			StatusStruct: nil,
		}
	}

	return resp, nil
}

func StartMockDVMService(listener net.Listener) *MockDVM {
	s := &MockDVM{
		execDelay: 100 * time.Millisecond,
	}

	server := grpc.NewServer()
	vm_grpc.RegisterVMModulePublisherServer(server, s)

	go func() {
		if err := server.Serve(listener); err != nil {
			fmt.Printf("MockDVM serve: %v\n", err)
		}
	}()
	s.server = server

	return s
}
