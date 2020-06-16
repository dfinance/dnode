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

	dvmCompilerSocketPath := path.Join(ct.UDSDir, dvmSocket)

	// Start DVM container
	dvmContainer, err := tests.NewDVMWithUDSTransport(ct.UDSDir, dvmSocket, dsSocket)
	require.NoError(t, err, "creating DVM container")
	require.NoError(t, dvmContainer.Start(5*time.Second), "staring DVM container")
	defer dvmContainer.Stop()

	// Wait for container to start up and register DVM socket address
	require.NoError(t, cliTester.WaitForFileExists(dvmCompilerSocketPath, 10*time.Second), "DVM gRPC server start")
	ct.SetVMCompilerAddressUDS(dvmCompilerSocketPath)

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
