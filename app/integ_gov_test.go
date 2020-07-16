// +build integ

package app

import (
	"encoding/hex"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	"github.com/dfinance/dnode/x/common_vm"
)

const govUpdModuleV1 = `
	address 0x1 {
	module Foo {
	    resource struct U64 {val: u64}
	    resource struct Address {val: address}

	    public fun store_u64(sender: &signer) {
			let value = U64 {val: 1};
	        move_to<U64>(sender, value);
	    }
	}
	}
`

const govUpdModuleV2 = `
	address 0x1 {
	module Foo {
	    resource struct U64 {val: u64}
	    resource struct Address {val: address}

	    public fun store_u64(sender: &signer) {
			let value = U64 {val: 2};
	        move_to<U64>(sender, value);
	    }
	}
	}
`

const govScript = `
	script {
		use 0x1::Foo;
		fun main(account: &signer) {
   			Foo::store_u64(account);
		}
	}
`

func TestIntegGov_StdlibUpdate(t *testing.T) {
	const (
		moduleAddr = "0000000000000000000000000000000000000001"
		modulePath = "0058e1e3e2f8edf7df0c4b1ab8c1c8ec661b3210b29c85b1449ac6214c6476b0e8"
	)

	ct := cliTester.New(
		t,
		true,
		cliTester.DaemonLogLevelOption("x/vm/dsserver:info,x/vm:info,x/gov:info,main:info,state:info,*:error"),
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
		movePath := path.Join(ct.Dirs.RootDir, fileName+".move")
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
		ct.QueryVmCompile(scriptMovePath, scriptBytecodePath, senderAddr).CheckFailedWithErrorSubstring("not found")
	}

	// Compile modules
	{
		ct.QueryVmCompile(moduleV1MovePath, moduleV1BytecodePath, senderAddr).CheckSucceeded()
		ct.QueryVmCompile(moduleV2MovePath, moduleV2BytecodePath, senderAddr).CheckSucceeded()
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
		ct.SubmitAndConfirmProposal(tx, true)
	}

	// Check module added and script works now
	{
		t.Log("Compiling script")
		ct.QueryVmCompile(scriptMovePath, scriptBytecodePath, senderAddr).CheckSucceeded()
		ct.TxVmExecuteScript(senderAddr, scriptBytecodePath).CheckSucceeded()
	}

	// Save module writeSet to compare later (if fails, 100% Libra path has changed again)
	moduleV1WriteSet := ""
	{
		q, writeSet := ct.QueryVmData(moduleAddr, modulePath)
		q.CheckSucceeded()
		require.NotEmpty(t, writeSet, "moduleV1 writeSet is empty")

		moduleV1WriteSet = writeSet.Value
	}

	// Add DVM stdlib update proposal for module version 2
	{
		tx := ct.TxVmStdlibUpdateProposal(senderAddr, moduleV2BytecodePath, "http://ya.ru", "Foo module V2 added", -1, config.GovMinDeposit)
		ct.SubmitAndConfirmProposal(tx, true)
	}

	// Check module writeSet changed
	{
		q, writeSet := ct.QueryVmData(moduleAddr, modulePath)
		q.CheckSucceeded()
		require.NotEmpty(t, writeSet, "moduleV2 writeSet is empty")

		moduleV2WriteSet := writeSet.Value
		require.NotEqual(t, moduleV1WriteSet, moduleV2WriteSet)
	}
}

func TestIntegGov_RegisterCurrency(t *testing.T) {
	ct := cliTester.New(
		t,
		true,
		cliTester.DaemonLogLevelOption("x/currencies_register:info,x/gov:info,main:info,state:info,*:error"),
	)
	defer ct.Close()

	wsStop, wsChs := ct.CheckWSsSubscribed(false, "TestIntegGov_RegisterCurrency", []string{"message.module='ccstorage'"}, 10)
	defer wsStop()
	go cliTester.PrintEvents(t, wsChs, "ccstorage")

	senderAddr := ct.Accounts["validator1"].Address

	// New currency info
	crDenom := "tst"
	crDecimals := uint8(8)
	crBalancePathHex := "A1A2A3A4A5A6A7A8A9ABACADAEAFB1B2B3B4B5B6B7B8B9BABBBCBDBEBFC1C2C3C4"
	crInfoPathHex    := "0102030405060708090A0B0C0D0E0FA1A2A3A4A5A6A7A8A9AAABACADAEAFB1B2B3"

	// Check invalid arguments for AddCurrencyProposal Tx
	{
		// invalid from
		{
			tx := ct.TxCCAddCurrencyProposal("invalid_from", crDenom, crBalancePathHex, crInfoPathHex, crDecimals, config.GovMinDeposit)
			tx.CheckFailedWithErrorSubstring("keyring")
		}

		// invalid denom
		{
			tx := ct.TxCCAddCurrencyProposal(senderAddr, "invalid1", crBalancePathHex, crInfoPathHex, crDecimals, config.GovMinDeposit)
			tx.CheckFailedWithErrorSubstring("denom")
		}

		// invalid path
		{
			tx1 := ct.TxCCAddCurrencyProposal(senderAddr, crDenom, "zzvv", crInfoPathHex, crDecimals, config.GovMinDeposit)
			tx1.CheckFailedWithErrorSubstring("vmBalancePathHex")
			tx2 := ct.TxCCAddCurrencyProposal(senderAddr, crDenom, crBalancePathHex, "abc", crDecimals, config.GovMinDeposit)
			tx2.CheckFailedWithErrorSubstring("vmInfoPathHex")
		}

		// invalid decimals
		{
			tx := ct.TxCCAddCurrencyProposal(senderAddr, crDenom, crBalancePathHex, crInfoPathHex, crDecimals, config.GovMinDeposit)
			tx.ChangeCmdArg("8", "abc")
			tx.CheckFailedWithErrorSubstring("decimals")
		}
	}

	// Add proposal
	{
		tx := ct.TxCCAddCurrencyProposal(senderAddr, crDenom, crBalancePathHex, crInfoPathHex, crDecimals, config.GovMinDeposit)
		ct.SubmitAndConfirmProposal(tx, false)
	}

	// Check currency added
	{
		req, currency := ct.QueryCurrenciesCurrency(crDenom)
		req.CheckSucceeded()

		require.Equal(t, crDenom, string(currency.Denom))
		require.Equal(t, crDecimals, currency.Decimals)
		require.True(t, currency.Supply.IsZero())
	}

	// Check writeSet is stored
	{
		q, writeSet := ct.QueryVmData(hex.EncodeToString(common_vm.StdLibAddress), crInfoPathHex)
		q.CheckSucceeded()
		require.NotEmpty(t, writeSet)
	}
}
