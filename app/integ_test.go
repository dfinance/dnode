// +build integ

package app

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

func Test_ConsensusFailure(t *testing.T) {
	const script = `
		script {
			use 0x0::Account;
			use 0x0::DFI;
			
			fun main(recipient: address, amount: u128) {
				Account::pay_from_sender<DFI::T>(recipient, amount);
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

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:" + compilerPort)

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
			use 0x0::Transaction;
			use 0x0::DFI;

			fun main() {
				Account::can_accept<DFI::T>(Transaction::sender());
			}
	}
`

	//t.Parallel()
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

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:" + compilerPort)

	// Start VM runtime
	runtimeContainer, err := tests.NewVMRuntimeContainerWithNetTransport(ct.VmConnectPort, ct.VmListenPort)
	require.NoError(t, err, "creating VM runtime container")

	require.NoError(t, runtimeContainer.Start(5*time.Second), "staring VM runtime container")
	defer runtimeContainer.Stop()

	senderAddr := ct.Accounts["validator1"].Address
	mvirPath := path.Join(ct.RootDir, "script.mvir")
	compiledPath := path.Join(ct.RootDir, "script.json")

	// Create .mvir script file
	mvirFile, err := os.Create(mvirPath)
	require.NoError(t, err, "creating script file")
	_, err = mvirFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, mvirFile.Close(), "close script file")

	// Compile .mvir script file
	ct.QueryVmCompileScript(mvirPath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}

// Test is skipped as it should be used for dnode <-> dvm communication over UDS debug locally (with DVM binaries).
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
			use 0x0::Transaction;
			use 0x0::DFI;

			fun main() {
				Account::can_accept<DFI::T>(Transaction::sender());
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
