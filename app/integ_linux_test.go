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

// Only runs on Linux based host machine (Docker for Mac can't work with UNIX sockets unfortunately).
func Test_VMCommunicationUDS(t *testing.T) {
	const (
		dsSocket         = "ds.sock"
		vmCompilerSocket = "vm_comp.sock"
		vmExecutorSocket = "vm_exec.sock"
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

	//t.Parallel()
	ct := cliTester.New(
		t,
		true,
		cliTester.VMConnectionSettings(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressUDS(dsSocket, vmExecutorSocket),
	)
	defer ct.Close()

	// Start VM compiler
	compilerContainer, err := tests.NewVMCompilerContainerWithUDSTransport(ct.UDSDir, dsSocket, vmCompilerSocket)
	require.NoError(t, err, "VM compiler container creation")

	require.NoError(t, compilerContainer.Start(5*time.Second), "VM compiler container start")
	defer compilerContainer.Stop()

	time.Sleep(60 * time.Second)

	// Wait for container to start up and register compiler socket address
	vmCompilerSocketPath := path.Join(ct.UDSDir, vmCompilerSocket)
	require.NoError(t, cliTester.WaitForFileExists(vmCompilerSocketPath, 10 * time.Second), "VM compiler container bool")
	ct.SetVMCompilerAddressUDS(vmCompilerSocketPath)

	senderAddr := ct.Accounts["validator1"].Address
	mvirPath := path.Join(ct.RootDir, "script.mvir")
	compiledPath := path.Join(ct.RootDir, "script.json")

	// Create .mvir script file
	mvirFile, err := os.Create(mvirPath)
	require.NoError(t, err, "creating script file")
	_, err = mvirFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, mvirFile.Close(), "close script file")

	//for {
	//	time.Sleep(1 * time.Second)
	//	t.Logf("I'm alive")
	//}

	// Compile .mvir script file
	ct.QueryVmCompileScript(mvirPath, compiledPath, senderAddr).CheckSucceeded()
}
