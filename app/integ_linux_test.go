// +build integ,linux

package app

import (
	"os"
	"path"
	"testing"

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
		cliTester.VMCommunicationOption(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressUDSOption(dsSocket, dvmSocket),
	)
	defer ct.Close()

	// Start DVM container and set compiler address
	dvmStop := tests.LaunchDVMWithUDSTransport(t, ct.Dirs.UDSDir, dvmSocket, dsSocket, false)
	defer dvmStop()

	ct.SetVMCompilerAddressUDS(path.Join(ct.Dirs.UDSDir, dvmSocket))

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.json")

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
