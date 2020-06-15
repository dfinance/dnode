// +build integ,linux

package app

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

func Test_VMCommunicationUDSOverDocker(t *testing.T) {
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

	// Start VM compiler (docker)
	compilerContainer, err := tests.NewVMCompilerContainerWithUDSTransport(ct.UDSDir, dsSocket, vmCompilerSocket)
	require.NoError(t, err, "creating VM compiler container")
	require.NoError(t, compilerContainer.Start(5*time.Second), "staring VM compiler container")
	defer compilerContainer.Stop()

	// Wait for compiler to start up and register compiler socket address
	require.NoError(t, cliTester.WaitForFileExists(vmCompilerSocketPath, 10*time.Second), "VM compiler gRPC server start")
	ct.SetVMCompilerAddressUDS(vmCompilerSocketPath)

	// Start VM runtime
	runtimeContainer, err := tests.NewVMRuntimeContainerWithUDSTransport(ct.UDSDir, dsSocket, vmRuntimeSocket)
	require.NoError(t, err, "creating VM runtime container")
	require.NoError(t, runtimeContainer.Start(5*time.Second), "staring VM runtime container")
	defer runtimeContainer.Stop()

	// Wait for runtime to start up
	require.NoError(t, cliTester.WaitForFileExists(vmRuntimeSocketPath, 10*time.Second), "VM runtime gRPC server start")

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.RootDir, "script.move")
	compiledPath := path.Join(ct.RootDir, "script.json")

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
