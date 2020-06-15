// +build integ

package app

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

type MockDVM struct {
	server        *grpc.Server
	failExecution bool
	failResponse  bool
	execDelay     time.Duration
}

func (s *MockDVM) SetExecutionFail() { s.failExecution = true }
func (s *MockDVM) SetExecutionOK()   { s.failExecution = false }
func (s *MockDVM) SetResponseFail()  { s.failResponse = true }
func (s *MockDVM) SetResponseOK()    { s.failResponse = false }
func (s *MockDVM) SetExecutionDelay(dur time.Duration) {
	s.execDelay = dur
}
func (s *MockDVM) Stop() {
	if s.server != nil {
		s.server.Stop()
	}
}

func (s *MockDVM) ExecuteContracts(ctx context.Context, req *vm_grpc.VMExecuteRequest) (*vm_grpc.VMExecuteResponses, error) {
	if s.failExecution {
		return nil, grpcStatus.Errorf(codes.Internal, "failing gRPC execution")
	}

	resp := &vm_grpc.VMExecuteResponses{}
	if !s.failResponse {
		resp.Executions = []*vm_grpc.VMExecuteResponse{
			{
				WriteSet:     nil,
				Events:       nil,
				GasUsed:      1,
				Status:       vm_grpc.ContractStatus_Discard,
				StatusStruct: nil,
			},
		}
	}

	return resp, nil
}

func StartMockDVMService(listener net.Listener) *MockDVM {
	s := &MockDVM{
		execDelay: 100 * time.Millisecond,
	}

	server := grpc.NewServer()
	vm_grpc.RegisterVMServiceServer(server, s)

	go func() {
		server.Serve(listener)
	}()
	s.server = server

	return s
}

func Test_ConsensusFailure(t *testing.T) {
	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;
			
			fun main(account: &signer, recipient: address, amount: u128) {
				Account::pay_from_sender<DFI::T>(account, recipient, amount);
			}
		}
`

	//t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	//ct.SetVMCompilerAddressNet("rpc.demo.wings.toys:50053")

	// Start VM compiler
	compilerContainer, compilerPort, err := tests.NewVMCompilerContainerWithNetTransport(ct.VmListenPort)
	require.NoError(t, err, "creating VM compiler container")

	require.NoError(t, compilerContainer.Start(5*time.Second), "staring VM compiler container")
	defer compilerContainer.Stop()

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:"+compilerPort, false)

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompileScript(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	// Should panic as there is no local VM running
	ct.TxVmExecuteScript(senderAddr, compiledPath, senderAddr, "100").DisableBroadcastMode().CheckSucceeded()

	// Check CONSENSUS FAILURE did occur
	{
		consensusFailure := false
		for i := 0; i < 10; i++ {
			if ct.DaemonLogsContain("CONSENSUS FAILURE") {
				consensusFailure = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		require.True(t, consensusFailure, "CONSENSUS FAILURE not occurred")
	}

	// Check restarted application panics
	{
		ct.RestartDaemon(false, false)

		retCode, daemonLogs := ct.CheckDaemonStopped(2 * time.Second)

		require.NotZero(t, retCode, "daemon exitCode")
		require.Contains(t, strings.Join(daemonLogs, ","), "panic", "daemon didn't panic")
	}
}

func Test_VMExecuteScript(t *testing.T) {
	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
	}
`

	ct := cliTester.New(
		t,
		true,
		cliTester.VMConnectionSettings(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressNet("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Start VM compiler
	compilerContainer, compilerPort, err := tests.NewVMCompilerContainerWithNetTransport(ct.VmListenPort)
	require.NoError(t, err, "creating VM compiler container")

	require.NoError(t, compilerContainer.Start(5*time.Second), "staring VM compiler container")
	defer compilerContainer.Stop()

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:"+compilerPort, false)

	// Start VM runtime
	runtimeContainer, err := tests.NewVMRuntimeContainerWithNetTransport(ct.VmConnectPort, ct.VmListenPort)
	require.NoError(t, err, "creating VM runtime container")

	require.NoError(t, runtimeContainer.Start(5*time.Second), "staring VM runtime container")
	defer runtimeContainer.Stop()

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.json")

	// Create .moe script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompileScript(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

func Test_VMRequestRetry(t *testing.T) {
	// TODO: Test should be rewritten as its success / failure is Moon phase dependant (not repeatable)
	t.Skip()

	const (
		dsSocket      = "ds.sock"
		mockDVMSocket = "mock_dvm.sock"
	)

	ct := cliTester.New(
		t,
		true,
		cliTester.VMConnectionSettings(100, 500, 10),
		cliTester.VMCommunicationBaseAddressUDS(dsSocket, mockDVMSocket),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	mockDVMSocketPath := path.Join(ct.UDSDir, mockDVMSocket)
	mockDVMListener, err := helpers.GetGRpcNetListener("unix://" + mockDVMSocketPath)
	require.NoError(t, err, "creating MockDVM listener")

	mockDvm := StartMockDVMService(mockDVMListener)
	defer mockDvm.Stop()
	require.NoError(t, cliTester.WaitForFileExists(mockDVMSocketPath, 1*time.Second), "MockDVM start failed")

	// Create fake .mov file
	modulePath := path.Join(ct.RootDir, "fake.json")
	moduleContent := []byte("{ \"code\": \"00\" }")
	require.NoError(t, ioutil.WriteFile(modulePath, moduleContent, 0644), "creating fake script file")

	wg := sync.WaitGroup{}
	vmDeployDoneCh := make(chan bool)

	// Spam REST requests while dnode is stucked on VM request
	// Stop only once VM is done, that ensures routines were parallel
	{
		wg.Add(1)
		go func() {
			defer wg.Done()

			t.Logf("RestRequest: start")
			for {
				req, _ := ct.RestQueryOracleAssets()
				req.SetTimeout(1000 * time.Millisecond)
				req.CheckSucceeded()
				t.Logf("RestRequest: ok")

				select {
				case <-vmDeployDoneCh:
					t.Logf("RestRequest: stop")
					return
				default:
					time.Sleep(100 * time.Millisecond)
				}
			}
		}()
	}

	// Execute .json module file
	// That should take some time and when "done" we close the channel to stop the first routine
	{
		mockDvm.SetExecutionDelay(3 * time.Second)
		senderAddr := ct.Accounts["validator1"].Address

		wg.Add(1)
		go func() {
			defer func() {
				close(vmDeployDoneCh)
				wg.Done()
			}()

			t.Logf("VMDeploy: start")
			ct.TxVmDeployModule(senderAddr, modulePath).CheckSucceeded()
			t.Logf("VMDeploy: done")
		}()
	}

	wg.Wait()
}

// Test is skipped: should be used for dnode <-> dvm (compiler / runtime) communication over UDS debug locally (with DVM binaries).
func Test_VMCommunicationUDSOverBinary(t *testing.T) {
	t.Skip()

	const (
		dsSocket         = "ds.sock"
		vmCompilerSocket = "compiler.sock"
		vmRuntimeSocket  = "runtime.sock"
	)

	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
	}
`

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.VMConnectionSettings(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressUDS(dsSocket, vmRuntimeSocket),
	)
	defer ct.Close()

	vmCompilerSocketPath := path.Join(ct.UDSDir, vmCompilerSocket)
	vmRuntimeSocketPath := path.Join(ct.UDSDir, vmRuntimeSocket)
	dsSocketPath := path.Join(ct.UDSDir, dsSocket)

	// Start VM compiler (sub-process)
	compilerCmd := cliTester.NewCLICmd(t, "compiler", "ipc:/"+vmCompilerSocketPath, "ipc:/"+dsSocketPath)
	compilerCmd.Start(t, false)
	defer compilerCmd.Stop()

	// Wait for compiler to start up and register compiler socket address
	require.NoError(t, cliTester.WaitForFileExists(vmCompilerSocketPath, 10*time.Second), "VM compiler gRPC server start")
	ct.SetVMCompilerAddressUDS(vmCompilerSocketPath)

	// Start VM runtime (sub-process)
	runtimeCmd := cliTester.NewCLICmd(t, "dvm", "ipc:/"+vmRuntimeSocketPath, "ipc:/"+dsSocketPath)
	runtimeCmd.Start(t, false)
	defer runtimeCmd.Stop()

	// Wait for runtime to start up
	require.NoError(t, cliTester.WaitForFileExists(vmRuntimeSocketPath, 10*time.Second), "VM runtime gRPC server start")

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompileScript(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

// Test is skipped: should be used for dnode <-> dvm (uni-binary) communication over TCP debug locally (with DVM binaries).
func Test_VMCommunicationTCPOverUniBinary(t *testing.T) {
	t.Skip()

	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
	}
`

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.VMConnectionSettings(50, 1000, 100),
	)
	defer ct.Close()

	// Start DVM compiler / runtime (sub-process) abd register compiler
	dvmAddr, dsAddr := "127.0.0.1:"+ct.VmConnectPort, "127.0.0.1:"+ct.VmListenPort
	dvmCmd := cliTester.NewCLICmd(t, "dvm", "http://"+dvmAddr, "http://"+dsAddr)
	dvmCmd.Start(t, true)
	defer dvmCmd.Stop()
	time.Sleep(1 * time.Second)

	// We skip TCP port ping as DVM doesn't open a new port
	ct.SetVMCompilerAddressNet(dvmAddr, true)

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompileScript(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

// Test is skipped: should be used for dnode <-> dvm (uni-binary) communication over UDS debug locally (with DVM binaries).
func Test_VMCommunicationUDSOverUniBinary(t *testing.T) {
	t.Skip()

	const (
		dsSocket  = "ds.sock"
		dvmSocket = "dvm.sock"
	)

	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;

			fun main(account: &signer) {
				let dfi = Account::withdraw_from_sender<DFI::T>(account, 1);
				Account::deposit_to_sender<DFI::T>(account, dfi);
			}
	}
`

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.VMConnectionSettings(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressUDS(dsSocket, dvmSocket),
	)
	defer ct.Close()

	dvmSocketPath := path.Join(ct.UDSDir, dvmSocket)
	dsSocketPath := path.Join(ct.UDSDir, dsSocket)

	// Start DVM compiler / runtime (sub-process) abd register compiler
	dvmCmd := cliTester.NewCLICmd(t, "dvm", "ipc:/"+dvmSocketPath, "ipc:/"+dsSocketPath)
	dvmCmd.Start(t, true)
	defer dvmCmd.Stop()
	time.Sleep(1 * time.Second)

	ct.SetVMCompilerAddressUDS(dvmSocketPath)

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompileScript(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}
