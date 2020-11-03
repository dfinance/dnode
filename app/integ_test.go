// +build integ

package app

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/helpers/tests/mockdvm"
	testUtils "github.com/dfinance/dnode/helpers/tests/utils"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
)

// Test dnode crash on VM Tx failure
func TestInteg_ConsensusFailure(t *testing.T) {
	const script = `
		script {
			use 0x1::Account;
			use 0x1::XFI;
			
			fun main(account: &signer, recipient: address, amount: u128) {
				Account::pay_from_sender<XFI::T>(account, recipient, amount);
			}
		}
	`

	ct := cliTester.New(t, false)
	defer ct.Close()

	// Start DVM compiler container (runtime also, but we don't want for dnode to connect to DVM runtime)
	_, vmCompilerPort, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for DVM compiler port")
	compilerStop := tests.LaunchDVMWithNetTransport(t, vmCompilerPort, ct.VMConnection.ListenPort, false)
	defer compilerStop()

	ct.SetVMCompilerAddressNet("tcp://127.0.0.1:"+vmCompilerPort, false)

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")
	// Compile .move script file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()
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

		retCode, daemonLogs := ct.CheckDaemonStopped(5 * time.Second)

		require.NotZero(t, retCode, "daemon exitCode")
		require.Contains(t, strings.Join(daemonLogs, ","), "panic", "daemon didn't panic")
	}
}

// Test Move compile and execute Move script with arg via CLI interface.
func TestIntegVM_ExecuteScriptViaCLI(t *testing.T) {
	const script = `
		script {
			use 0x1::Account;
			use 0x1::XFI;

			fun main(account: &signer, amount: u128) {
				let xfi = Account::withdraw_from_sender<XFI::T>(account, amount);
				Account::deposit_to_sender<XFI::T>(account, xfi);
			}
		}
	`

	const multiScript = `
		script {
			use 0x1::Account;
			use 0x1::XFI;

			fun main(account: &signer, amount: u128) {
				let xfi = Account::withdraw_from_sender<XFI::T>(account, amount);
				Account::deposit_to_sender<XFI::T>(account, xfi);
			}
		}

		script {
			use 0x1::Event;
			
			fun main(account: &signer, a: u64) {
				Event::emit<u64>(account, a);
			}
		}
	`

	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

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
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Check compile query cmd with invalid inputs
	{
		// invalid source path
		{
			r := ct.QueryVmCompile("./invalid.json", compiledPath, senderAddr)
			r.CheckFailedWithErrorSubstring("moveFile")
		}
		// invalid account
		{
			r := ct.QueryVmCompile(movePath, compiledPath, "invalid")
			r.CheckFailedWithErrorSubstring("account")
		}
	}

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath, "100").CheckSucceeded()

	// Check execute Tx cmd with invalid inputs
	{
		// invalid from flag
		{
			r := ct.TxVmExecuteScript("invalid", compiledPath, "100")
			r.CheckFailedWithErrorSubstring("from")
		}
		// invalid source path
		{
			r := ct.TxVmExecuteScript(senderAddr, "./invalid.json", "100")
			r.CheckFailedWithErrorSubstring("moveFile")
		}
		// invalid args len (no args)
		{
			r := ct.TxVmExecuteScript(senderAddr, compiledPath)
			r.CheckFailedWithErrorSubstring("length mismatch")
		}
		// invalid args len (more than needed)
		{
			r := ct.TxVmExecuteScript(senderAddr, compiledPath, "100", "200")
			r.CheckFailedWithErrorSubstring("length mismatch")
		}
		// invalid args (wrong type)
		{
			r := ct.TxVmExecuteScript(senderAddr, compiledPath, "true")
			r.CheckFailedWithErrorSubstring("true")
		}
	}

	// Check multi script compile and execution
	{
		movePath := path.Join(ct.Dirs.RootDir, "mscript.move")
		compiledPath := path.Join(ct.Dirs.RootDir, "mscript.json")

		// Create .move script file
		moveFile, err := os.Create(movePath)
		require.NoError(t, err, "creating script file")
		_, err = moveFile.WriteString(multiScript)
		require.NoError(t, err, "write script file")
		require.NoError(t, moveFile.Close(), "close script file")

		// Compile .move script file
		ct.QueryVmCompile(movePath, compiledPath, senderAddr).
			CheckSucceeded()

		// Execute compiled script file
		ct.TxVmExecuteScript(senderAddr, compiledPath, "100").
			CheckFailedWithErrorSubstring("wrong quantity of items")
	}

}

// Test Move compile and execute Move script with arg via REST interface.
func TestIntegVM_ExecuteScriptViaREST(t *testing.T) {
	const script = `
		script {
			use 0x1::Account;
			use 0x1::XFI;

			fun main(account: &signer, amount: u128) {
				let xfi = Account::withdraw_from_sender<XFI::T>(account, amount);
				Account::deposit_to_sender<XFI::T>(account, xfi);
			}
		}
	`

	const dualScript = `
		script {
			fun main(x: u64, y: u64) {
				x - y;
			}
		}

		script {
			fun main(a: u64, b: u64) {
				a + b;
			}
		}
	`

	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	senderName := "validator1"
	senderAddress := ct.Accounts[senderName].Address

	// Compile 2 scripts in the one request
	{
		r, resp := ct.RestQueryVMCompile(senderAddress, dualScript)
		r.CheckSucceeded()
		require.NotEmpty(t, resp)
		require.Len(t, *resp, 2)
	}

	r, resp := ct.RestQueryVMCompile(senderAddress, script)
	r.CheckSucceeded()

	require.NotEmpty(t, resp)
	require.Len(t, *resp, 1)
	byteCode := (*resp)[0].Code

	// Check compile endpoint with invalid inputs
	{
		// invalid account
		{
			r, _ := ct.RestQueryVMCompile("invalid", script)
			r.CheckFailed(400, nil)
		}
	}

	// Execute script
	{
		arg := "100"
		q, stdTx := ct.RestQueryVMExecuteScriptStdTx(senderName, byteCode, "", arg)
		q.CheckSucceeded()

		// verify stdTx
		{
			scriptArg, err := vm_client.NewU128ScriptArg(arg)
			require.NoError(t, err)

			code, err := hex.DecodeString(byteCode)
			require.NoError(t, err)

			require.Len(t, stdTx.Msgs, 1)
			require.IsType(t, vm.MsgExecuteScript{}, stdTx.Msgs[0])
			executeMsg := stdTx.Msgs[0].(vm.MsgExecuteScript)
			require.EqualValues(t, code, executeMsg.Script)
			require.Len(t, executeMsg.Args, 1)
			require.EqualValues(t, scriptArg, executeMsg.Args[0])
			require.Equal(t, ct.Accounts[senderName].Address, executeMsg.Signer.String())
		}

		// run Tx
		r, _ := ct.NewRestStdTxRequest(senderName, *stdTx, false)
		r.CheckSucceeded()
	}

	// Check execute script endpoint with invalid inputs
	{
		// invalid code (non-hex string)
		{
			q, _ := ct.RestQueryVMExecuteScriptStdTx(senderName, "zxy", "", "100")
			q.CheckFailed(400, nil)
		}
		// invalid args len (no args)
		{
			q, _ := ct.RestQueryVMExecuteScriptStdTx(senderName, byteCode, "")
			q.CheckFailed(400, nil)
		}
		// invalid args len (more than needed)
		{
			q, _ := ct.RestQueryVMExecuteScriptStdTx(senderName, byteCode, "", "100", "200")
			q.CheckFailed(400, nil)
		}
		// invalid args (wrong type)
		{
			q, _ := ct.RestQueryVMExecuteScriptStdTx(senderName, byteCode, "", "true")
			q.CheckFailed(400, nil)
		}
	}
}

// Deploy multiple Move modules via CLI interface.
func TestIntegVM_DeployModuleViaCLI(t *testing.T) {
	const module = `
		address 0x1 {
			module Foo {
				resource struct U64 {val: u64}
				public fun store_u64(sender: &signer) {
					let value = U64 {val: 1};
					move_to<U64>(sender, value);
				}
			}

			module Bar {
				public fun sub(a: u64, b: u64): u64 {
					a - b
				}
			}
		}
	`

	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "module.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "module.json")

	// Create .move module file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating module file")
	_, err = moveFile.WriteString(module)
	require.NoError(t, err, "write module file")
	require.NoError(t, moveFile.Close(), "close module file")

	// Compile .move module file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	compiledPathContent, _ := ioutil.ReadFile(compiledPath)
	var modules vm_client.CompiledItems
	jsonErr := json.Unmarshal(compiledPathContent, &modules)
	require.NoError(t, jsonErr, "parse move file")
	require.Len(t, modules, 2)

	// Execute .json script file
	ct.TxVmDeployModule(senderAddr, compiledPath).CheckSucceeded()

	// Check deploy contract Tx cmd with invalid inputs
	{
		// invalid from flag
		{
			r := ct.TxVmDeployModule("invalid", compiledPath)
			r.CheckFailedWithErrorSubstring("from")
		}
		// invalid source path
		{
			r := ct.TxVmDeployModule(senderAddr, "./invalid.json")
			r.CheckFailedWithErrorSubstring("moveFile")
		}
	}
}

// Deploy Move module via REST interface.
func TestIntegVM_DeployModuleViaREST(t *testing.T) {
	const module = `
		address 0x1 {
			module Foo {
				resource struct U64 {val: u64}
				
				public fun store_u64(sender: &signer) {
					let value = U64 {val: 1};
					move_to<U64>(sender, value);
				}
			}

			module Bar {
				public fun sub(a: u64, b: u64): u64 {
					a - b
				}
			}
		}
	`

	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	senderName := "validator1"
	senderAddress := ct.Accounts[senderName].Address

	// Compile module
	var byteCode []string
	{
		r, resp := ct.RestQueryVMCompile(senderAddress, module)
		r.CheckSucceeded()

		require.NotEmpty(t, resp)

		for _, item := range *resp {
			byteCode = append(byteCode, item.Code)
		}
	}

	// Deploy script
	{
		require.Equal(t, 2, len(byteCode))
		q, stdTx := ct.RestQueryVMPublishModuleStdTx(senderName, byteCode, "")
		q.CheckSucceeded()

		// verify stdTx
		{
			code1, err := hex.DecodeString(byteCode[0])
			code2, err := hex.DecodeString(byteCode[1])
			require.NoError(t, err)

			require.Len(t, stdTx.Msgs, 1)
			require.IsType(t, vm.MsgDeployModule{}, stdTx.Msgs[0])
			deployMsg := stdTx.Msgs[0].(vm.MsgDeployModule)
			require.EqualValues(t, []vm.Contract{code1, code2}, deployMsg.Module)
			require.Equal(t, ct.Accounts[senderName].Address, deployMsg.Signer.String())
		}

		// run Tx
		r, _ := ct.NewRestStdTxRequest(senderName, *stdTx, false)
		r.CheckSucceeded()
	}

	// Check deploy module endpoint with invalid inputs
	{
		// invalid code (non-hex string)
		{
			q, _ := ct.RestQueryVMPublishModuleStdTx(senderName, []string{"zxy"}, "")
			q.CheckFailed(400, nil)
		}
	}
}

// Tries execute and deploy mixed content (module and script) in one file
func TestIntegVM_DeployModuleAndScriptViaCLI(t *testing.T) {
	const module = `
		address 0x1 {
			module Foo {
				resource struct U64 {val: u64}
				public fun store_u64(sender: &signer) {
					let value = U64 {val: 1};
					move_to<U64>(sender, value);
				}
			}
		}

		script {
			use 0x1::Event;
			
			fun main(account: &signer, a: u64, b: u64) {
				let c = a + b;
				Event::emit<u64>(account, c);
			}
		}
	`

	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressNetOption("tcp://127.0.0.1"),
	)
	defer ct.Close()

	// Start DVM container
	dvmStop := tests.LaunchDVMWithNetTransport(t, ct.VMConnection.ConnectPort, ct.VMConnection.ListenPort, false)
	defer dvmStop()

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "module.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "module.json")

	// Create .move module file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating module file")
	_, err = moveFile.WriteString(module)
	require.NoError(t, err, "write module file")
	require.NoError(t, moveFile.Close(), "close module file")

	// Compile .move module file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	compiledPathContent, _ := ioutil.ReadFile(compiledPath)
	var modules vm_client.CompiledItems
	jsonErr := json.Unmarshal(compiledPathContent, &modules)
	require.NoError(t, jsonErr, "parse move file")
	require.Len(t, modules, 2)

	// Deploy .json script file
	ct.TxVmDeployModule(senderAddr, compiledPath).
		CheckFailedWithErrorSubstring("contains different code types")

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).
		CheckFailedWithErrorSubstring("contains wrong quantity of items")
}

// Test dnode <-> dvm request-retry mechanism.
func TestIntegVM_RequestRetry(t *testing.T) {
	// TODO: Test should be rewritten as its success / failure is Moon phase dependant (not repeatable)
	t.Skip("skipped as the current implementation is not stable")

	const (
		dsSocket      = "ds.sock"
		mockDVMSocket = "mock_dvm.sock"
	)

	ct := cliTester.New(
		t,
		true,
		cliTester.VMCommunicationOption(5, 100),
		cliTester.VMCommunicationBaseAddressUDSOption(dsSocket, mockDVMSocket),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	mockDVMSocketPath := path.Join(ct.Dirs.UDSDir, mockDVMSocket)
	mockDVMListener, err := helpers.GetGRpcNetListener("unix://" + mockDVMSocketPath)
	require.NoError(t, err, "creating MockDVM listener")

	mockDvm := mockdvm.StartMockDVMService(mockDVMListener)
	defer mockDvm.Stop()
	require.NoError(t, testUtils.WaitForFileExists(mockDVMSocketPath, 1*time.Second), "MockDVM start failed")

	// Create fake .mov file
	modulePath := path.Join(ct.Dirs.RootDir, "fake.json")
	moduleContent := []byte("{ \"code\": \"00\" }")
	require.NoError(t, ioutil.WriteFile(modulePath, moduleContent, 0644), "creating fake module file")

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

// Test is skipped: should be used for dnode <-> dvm (uni-binary) communication over UDS debug locally (with DVM binaries).
func TestIntegVM_CommunicationUDS(t *testing.T) {
	t.Skip("skipped as it is for local debug only")

	const (
		dsSocket  = "ds.sock"
		dvmSocket = "dvm.sock"
	)

	const script = `
		script {
			use 0x1::Account;
			use 0x1::XFI;

			fun main(account: &signer) {
				let xfi = Account::withdraw_from_sender<XFI::T>(account, 1);
				Account::deposit_to_sender<XFI::T>(account, xfi);
			}
		}
	`

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.VMCommunicationOption(5, 1000),
		cliTester.VMCommunicationBaseAddressUDSOption(dsSocket, dvmSocket),
	)
	defer ct.Close()

	// Start DVM compiler / runtime (sub-process) abd register compiler
	os.Setenv(tests.EnvDvmIntegUse, "binary")
	dvmStop := tests.LaunchDVMWithUDSTransport(t, ct.Dirs.UDSDir, dvmSocket, dsSocket, false)
	defer dvmStop()

	ct.SetVMCompilerAddressUDS(path.Join(ct.Dirs.UDSDir, dvmSocket))

	senderAddr := ct.Accounts["validator1"].Address
	movePath := path.Join(ct.Dirs.RootDir, "script.move")
	compiledPath := path.Join(ct.Dirs.RootDir, "script.move.json")

	// Create .move script file
	moveFile, err := os.Create(movePath)
	require.NoError(t, err, "creating script file")
	_, err = moveFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, moveFile.Close(), "close script file")

	// Compile .move script file
	ct.QueryVmCompile(movePath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	ct.TxVmExecuteScript(senderAddr, compiledPath).CheckSucceeded()
}
