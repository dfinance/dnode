// +build integ

package app

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

const govUpdModuleV1 = `
	address 0x0 {
	module Foo {
	    resource struct U64 {val: u64}
	    resource struct Address {val: address}

	    public fun store_u64() {
			let value = U64 {val: 1};
	        move_to_sender<U64>(value);
	    }
	}
	}
`

const govUpdModuleV2 = `
	address 0x0 {
	module Foo {
	    resource struct U64 {val: u64}
	    resource struct Address {val: address}

	    public fun store_u64() {
			let value = U64 {val: 2};
	        move_to_sender<U64>(value);
	    }
	}
	}
`

const govScript = `
	script {
		use 0x0::Foo;
		fun main() {
   			Foo::store_u64();
		}
	}
`

func Test_VmGovStdlibUpdate(t *testing.T) {
	ct := cliTester.New(
		t,
		true,
		cliTester.DaemonLogLevelOption("x/vm:info,x/gov:info,main:info,state:info,*:error"),
		cliTester.VMCommunicationOption(50, 1000, 100),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Capture all VM status events for ease the debug
	wsStop1, wsChan1 := ct.CheckWSSubscribed(false, "Test_VmGovStdlibUpdate1", "contract_status.status='keep'", 10)
	wsStop2, wsChan2 := ct.CheckWSSubscribed(false, "Test_VmGovStdlibUpdate2", "contract_status.status='discard'", 10)
	wsStop3, wsChan3 := ct.CheckWSSubscribed(false, "Test_VmGovStdlibUpdate3", "contract_status.status='error'", 10)
	defer wsStop1()
	defer wsStop2()
	defer wsStop3()

	go func() {
		for {
			select {
			case event, ok := <-wsChan1:
				if !ok {
					return
				}
				t.Logf("Got event (ch1): events: %v", event.Events)
			case event, ok := <-wsChan2:
				if !ok {
					return
				}
				t.Logf("Got event (ch2): events: %v", event.Events)
			case event, ok := <-wsChan3:
				if !ok {
					return
				}
				t.Logf("Got event (ch3): events: %v", event.Events)
			}
		}
	}()

	senderAddr := ct.Accounts["pos"].Address

	// Start DVM
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	createMovFile := func(fileName, code string) string {
		movePath := path.Join(ct.Dirs.RootDir, fileName + ".move")
		moveFile, err := os.Create(movePath)
		require.NoErrorf(t, err, "creating .move file for %s", fileName)
		_, err = moveFile.WriteString(code)
		require.NoErrorf(t, err, "write .move file for %s", fileName)
		require.NoErrorf(t, moveFile.Close(), "close .move file for %s", fileName)

		return movePath
	}

	scriptMovePath, scriptBytecodePath := createMovFile("script", govScript), path.Join(ct.Dirs.RootDir, "script.json")
	moduleV1MovePath, moduleV1BytecodePath := createMovFile("module_v1", govUpdModuleV1), path.Join(ct.Dirs.RootDir, "module_v1.json")
	moduleV2MovePath, moduleV2BytecodePath := createMovFile("module_v2", govUpdModuleV2), path.Join(ct.Dirs.RootDir, "module_v2.json")

	// Check script can't be compiled as module doesn't exist yet
	{
		ct.QueryVmCompileScript(scriptMovePath, scriptBytecodePath, senderAddr).CheckFailedWithErrorSubstring("not found")
	}

	// Compile modules
	{
		ct.QueryVmCompileModule(moduleV1MovePath, moduleV1BytecodePath, senderAddr).CheckSucceeded()
		ct.QueryVmCompileModule(moduleV2MovePath, moduleV2BytecodePath, senderAddr).CheckSucceeded()
	}

	// Check invalid arguments for StdlibUpdateProposal Tx
	{
		// invalid from
		{
			tx := ct.TxVmStdlibUpdateProposal("invalid_address", moduleV1BytecodePath, "http://ya.ru", "Desc", 50, config.GovMinDeposit)
			tx.CheckFailedWithErrorSubstring("keyring")
		}

		// invalid file path
		{
			tx := ct.TxVmStdlibUpdateProposal(senderAddr, "invalid_path", "http://ya.ru", "Desc", 50, config.GovMinDeposit)
			tx.CheckFailedWithErrorSubstring("mvFile")
		}

		// invalid blockHeight
		{
			tx1 := ct.TxVmStdlibUpdateProposal(senderAddr, moduleV1BytecodePath, "http://ya.ru", "Desc", 0, config.GovMinDeposit)
			tx1.CheckFailedWithErrorSubstring("height")

			tx2 := ct.TxVmStdlibUpdateProposal(senderAddr, moduleV1BytecodePath, "http://ya.ru", "Desc", 0, config.GovMinDeposit)
			tx2.ChangeCmdArg("0", "abc")
			tx2.CheckFailedWithErrorSubstring("ParseInt")
		}
	}

	// Add DVM stdlib update proposal for module version 1 (cover the min deposit)
	{
		tx := ct.TxVmStdlibUpdateProposal(senderAddr, moduleV1BytecodePath, "http://ya.ru", "Foo module V1 added", -1, config.GovMinDeposit)
		ct.SubmitAndConfirmProposal(tx)
	}

	// Check module added and script works now
	{
		t.Log("Compiling script")
		ct.QueryVmCompileScript(scriptMovePath, scriptBytecodePath, senderAddr).CheckSucceeded()
		ct.TxVmExecuteScript(senderAddr, scriptBytecodePath).CheckSucceeded()
	}

	// Save module writeSet to compare later (if fails, 100% Libra path has changed again)
	moduleV1WriteSet := ""
	{
		q, writeSet := ct.QueryVmData("0000000000000000000000000000000000000000", "00b3fafba7710d2bc614054f5cd7b53edc0da61bfae33cd7f1483fa50b5dd0029c")
		q.CheckSucceeded()

		moduleV1WriteSet = writeSet.Value
	}

	// Add DVM stdlib update proposal for module version 2
	{
		tx := ct.TxVmStdlibUpdateProposal(senderAddr, moduleV2BytecodePath, "http://ya.ru", "Foo module V2 added", -1, config.GovMinDeposit)
		ct.SubmitAndConfirmProposal(tx)
	}

	// Check module writeSet changed
	{
		q, writeSet := ct.QueryVmData("0000000000000000000000000000000000000000", "00b3fafba7710d2bc614054f5cd7b53edc0da61bfae33cd7f1483fa50b5dd0029c")
		q.CheckSucceeded()

		moduleV2WriteSet := writeSet.Value
		require.NotEqual(t, moduleV1WriteSet, moduleV2WriteSet)
	}
}
