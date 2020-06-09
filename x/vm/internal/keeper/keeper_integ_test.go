// +build integ

package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/OneOfOne/xxhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnodeConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle"
	compilerClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const sendScript = `
script {
	use 0x0::Account;
	use 0x0::Coins;
	use 0x0::DFI;
	
	fun main(recipient: address, dfi_amount: u128, eth_amount: u128, btc_amount: u128, usdt_amount: u128) {
		Account::pay_from_sender<DFI::T>(recipient, dfi_amount);
		Account::pay_from_sender<Coins::ETH>(recipient, eth_amount);
		Account::pay_from_sender<Coins::BTC>(recipient, btc_amount);
		Account::pay_from_sender<Coins::USDT>(recipient, usdt_amount);
	}
}
`

const mathModule = `
module Math {
    public fun add(a: u64, b: u64): u64 {
		a + b
    }
}
`

const mathScript = `
script {
	use 0x0::Event;
	use {{sender}}::Math;
	
	fun main(a: u64, b: u64) {
		let c = Math::add(a, b);
	
		let event_handle = Event::new_event_handle<u64>();
		Event::emit_event(&mut event_handle, c);
		Event::destroy_handle(event_handle);
	}
}
`

const oraclePriceScript = `
script {
	use 0x0::Event;
	use 0x0::Oracle;
	
	fun main(ticket: u64) {
		let price = Oracle::get_price(ticket);
	
		let event_handle = Event::new_event_handle<u64>();
		Event::emit_event(&mut event_handle, price);
		Event::destroy_handle(event_handle);
	}
}
`

const errorScript = `
script {
	use 0x0::Account;
	use 0x0::DFI;
	use 0x0::Transaction;
	use 0x0::Coins;
	use 0x0::Event;

	fun main(c: u64) {
		let a = Account::withdraw_from_sender<DFI::T>(523);
		let b = Account::withdraw_from_sender<Coins::BTC>(1);
	
	
		let event_handle = Event::new_event_handle<u64>();
		Event::emit_event(&mut event_handle, 10);
		Event::destroy_handle(event_handle);
	
		Transaction::assert(c == 1000, 122);
		Account::deposit_to_sender(a);
		Account::deposit_to_sender(b);
	}
}
`

// print events.
func printEvent(event sdk.Event, t *testing.T) {
	t.Logf("Event: %s\n", event.Type)
	for _, attr := range event.Attributes {
		t.Logf("%s = %s\n", attr.Key, attr.Value)
	}
}

// check that sub status exists.
func checkSubStatus(events sdk.Events, subStatus uint64, t *testing.T) {
	found := false

	for _, event := range events {
		if event.Type == types.EventTypeContractStatus {
			// find error
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttrKeySubStatus {
					require.EqualValues(t, attr.Value, []byte(strconv.FormatUint(subStatus, 10)), "wrong value for sub status")

					found = true
				}
			}
		}
	}

	require.True(t, found, "sub status not found")
}

// check script doesn't contains errors.
func checkNoErrors(events sdk.Events, t *testing.T) {
	for _, event := range events {
		if event.Type == types.EventTypeContractStatus {
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttrKeyStatus {
					if string(attr.Value) == types.StatusDiscard {
						printEvent(event, t)
						t.Fatalf("should not contains error event")
					}

					if string(attr.Value) == types.StatusError {
						printEvent(event, t)
						t.Fatalf("should not contains error event")
					}
				}
			}
		}
	}
}

// Test transfer of dfi between two accounts in dfi.
func TestKeeper_DeployContractTransfer(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// launch docker
	client, compiler, vm := launchDocker(input.dsPort, t)
	defer input.vk.CloseConnections()
	defer stopDocker(t, client, compiler)
	defer stopDocker(t, client, vm)

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)

	baseAmount := sdk.NewInt(1000)
	putCoins := sdk.NewCoins(
		sdk.NewCoin("dfi", baseAmount),
		sdk.NewCoin("eth", baseAmount),
		sdk.NewCoin("btc", baseAmount),
		sdk.NewCoin("usdt", baseAmount),
	)

	denoms := make([]string, 4)
	denoms[0] = "dfi"
	denoms[1] = "eth"
	denoms[2] = "btc"
	denoms[3] = "usdt"

	toSend := make(map[string]sdk.Int, 4)

	for i := 0; i < len(denoms); i++ {
		toSend[denoms[i]] = sdk.NewInt(100 - int64(i)*10)
	}

	acc1.SetCoins(putCoins)

	addr2 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc2 := input.ak.NewAccountWithAddress(input.ctx, addr2)

	input.ak.SetAccount(input.ctx, acc1)
	input.ak.SetAccount(input.ctx, acc2)

	// write write set.
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// wait for compiler
	err := waitStarted(client, compiler.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker compiler: %v", err)

	err = waitStarted(client, vm.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker vm: %v", err)

	// wait reachable compiler
	err = waitReachable(*vmCompiler, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to compiler server: %v", err)

	// wait reachable vm
	err = waitReachable(*vmAddress, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to vm server: %v", err)

	t.Logf("Compile send script")
	bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    sendScript,
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for send script: %v", err)

	// execute contract.
	args := make([]types.ScriptArg, 1)
	args[0] = types.ScriptArg{
		Value: addr2.String(),
		Type:  vm_grpc.VMTypeTag_Address,
	}

	for _, d := range denoms {
		args = append(args, types.ScriptArg{
			Value: toSend[d].String(),
			Type:  vm_grpc.VMTypeTag_U128,
		})
	}

	t.Logf("Execute send script")
	msgScript := types.NewMsgExecuteScript(addr1, bytecode, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	require.Contains(t, events, types.NewEventKeep())

	checkNoErrors(events, t)

	// check balance changes
	sender := input.ak.GetAccount(input.ctx, addr1)
	getCoins := sender.GetCoins()

	for _, got := range getCoins {
		require.Equalf(t, baseAmount.Sub(toSend[got.Denom]).String(), got.Amount.String(), "not equal for sender %s", got.Denom)
	}

	recipient := input.ak.GetAccount(input.ctx, addr2)
	recpCoins := recipient.GetCoins()

	for _, got := range recpCoins {
		require.Equalf(t, toSend[got.Denom].String(), got.Amount.String(), "not equal for recipient %s", got.Denom)
	}
}

// Test deploy module and use it.
func TestKeeper_DeployModule(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// launch docker
	client, compiler, vm := launchDocker(input.dsPort, t)
	defer stopDocker(t, client, vm)
	defer stopDocker(t, client, compiler)

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)

	input.ak.SetAccount(input.ctx, acc1)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// wait for compiler
	err := waitStarted(client, compiler.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker compiler: %v", err)

	err = waitStarted(client, vm.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker vm: %v", err)

	// wait reachable compiler
	err = waitReachable(*vmCompiler, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to compiler server: %v", err)

	// wait reachable vm
	err = waitReachable(*vmAddress, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to vm server: %v", err)

	bytecodeModule, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    mathModule,
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Module,
	})
	require.NoErrorf(t, err, "can't get code for math module: %v", err)

	msg := types.NewMsgDeployModule(addr1, bytecodeModule)
	err = msg.ValidateBasic()
	require.NoErrorf(t, err, "can't validate err: %v", err)

	ctx, writeCtx := input.ctx.CacheContext()
	err = input.vk.DeployContract(ctx, msg)
	require.NoErrorf(t, err, "can't deploy contract: %v", err)

	events := ctx.EventManager().Events()
	checkNoErrors(events, t)

	writeCtx()

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    strings.Replace(mathScript, "{{sender}}", addr1.String(), 1),
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't compiler script for math module: %v", err)

	args := make([]types.ScriptArg, 2)
	args[0] = types.ScriptArg{
		Value: "10",
		Type:  vm_grpc.VMTypeTag_U64,
	}
	args[1] = types.ScriptArg{
		Value: "100",
		Type:  vm_grpc.VMTypeTag_U64,
	}

	ctx, _ = input.ctx.CacheContext()
	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, args)
	err = input.vk.ExecuteScript(ctx, msgScript)
	require.NoError(t, err)

	events = ctx.EventManager().Events()
	require.Contains(t, events, types.NewEventKeep())

	checkNoErrors(events, t)

	require.Equal(t, events[1].Type, types.EventTypeMoveEvent, "script after execution doesn't contain event with amount")

	require.Len(t, events[1].Attributes, 4)
	require.EqualValues(t, events[1].Attributes[1].Key, types.AttrKeySequenceNumber)
	require.EqualValues(t, events[1].Attributes[1].Value, "0")
	require.EqualValues(t, events[1].Attributes[2].Key, types.AttrKeyType)
	require.EqualValues(t, events[1].Attributes[2].Value, types.VMTypeToStringPanic(vm_grpc.VMTypeTag_U64))
	require.EqualValues(t, events[1].Attributes[3].Key, types.AttrKeyData)

	uintBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(uintBz, uint64(110))

	require.EqualValues(t, events[1].Attributes[3].Value, "0x"+hex.EncodeToString(uintBz))
}

// Test oracle price return.
func TestKeeper_ScriptOracle(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// launch docker
	client, compiler, vm := launchDocker(input.dsPort, t)
	defer stopDocker(t, client, vm)
	defer stopDocker(t, client, compiler)

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)

	input.ak.SetAccount(input.ctx, acc1)

	assetCode := "eth_usdt"
	okInitParams := oracle.Params{
		Assets: oracle.Assets{
			oracle.Asset{
				AssetCode: assetCode,
				Oracles: oracle.Oracles{
					oracle.Oracle{
						Address: addr1,
					},
				},
				Active: true,
			},
		},
		Nominees: []string{addr1.String()},
		PostPrice: oracle.PostPriceParams{
			ReceivedAtDiffInS: 3600,
		},
	}

	input.ok.SetParams(input.ctx, okInitParams)
	input.ok.SetPrice(input.ctx, addr1, assetCode, sdk.NewInt(100), time.Now())
	input.ok.SetCurrentPrices(input.ctx)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// wait for compiler
	err := waitStarted(client, compiler.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker compiler: %v", err)

	err = waitStarted(client, vm.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker vm: %v", err)

	// wait reachable compiler
	err = waitReachable(*vmCompiler, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to compiler server: %v", err)

	// wait reachable vm
	err = waitReachable(*vmAddress, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to vm server: %v", err)

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    oraclePriceScript,
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for oracle script: %v", err)

	seed := xxhash.NewS64(0)
	_, err = seed.WriteString(strings.ToLower(assetCode))
	require.NoErrorf(t, err, "can't convert: %v", err)
	value := seed.Sum64()

	args := make([]types.ScriptArg, 1)
	args[0] = types.ScriptArg{
		Value: strconv.FormatUint(value, 10),
		Type:  vm_grpc.VMTypeTag_U64,
	}

	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	require.Contains(t, events, types.NewEventKeep())

	require.Len(t, events[1].Attributes, 4)
	require.EqualValues(t, events[1].Attributes[1].Key, types.AttrKeySequenceNumber)
	require.EqualValues(t, events[1].Attributes[1].Value, "0")
	require.EqualValues(t, events[1].Attributes[2].Key, types.AttrKeyType)
	require.EqualValues(t, events[1].Attributes[2].Value, types.VMTypeToStringPanic(vm_grpc.VMTypeTag_U64))
	require.EqualValues(t, events[1].Attributes[3].Key, types.AttrKeyData)

	bz := make([]byte, 8)

	binary.LittleEndian.PutUint64(bz, 100)
	require.EqualValues(t, events[1].Attributes[3].Value, "0x"+hex.EncodeToString(bz))

	checkNoErrors(events, t)
}

// Test oracle price return.
func TestKeeper_ErrorScript(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// launch docker
	client, compiler, vm := launchDocker(input.dsPort, t)
	defer stopDocker(t, client, vm)
	defer stopDocker(t, client, compiler)

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	coins := sdk.NewCoins(
		sdk.NewCoin("dfi", sdk.NewInt(1000000000000000)),
		sdk.NewCoin("btc", sdk.NewInt(1)),
	)

	acc1.SetCoins(coins)
	input.ak.SetAccount(input.ctx, acc1)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// wait for compiler
	err := waitStarted(client, compiler.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker compiler: %v", err)

	err = waitStarted(client, vm.ID, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to docker vm: %v", err)

	// wait reachable compiler
	err = waitReachable(*vmCompiler, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to compiler server: %v", err)

	// wait reachable vm
	err = waitReachable(*vmAddress, 5*time.Second)
	require.NoErrorf(t, err, "can't connect to vm server: %v", err)

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    errorScript,
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for error script: %v", err)

	args := make([]types.ScriptArg, 1)
	args[0] = types.ScriptArg{
		Value: strconv.FormatUint(10, 10),
		Type:  vm_grpc.VMTypeTag_U64,
	}

	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	require.Contains(t, events, types.NewEventKeep())
	for _, e := range events {
		printEvent(e, t)
	}
	checkSubStatus(events, 122, t)

	// first of all - check balance
	// then check that error still there
	// then check that no events there only error and keep status
	getAcc := input.ak.GetAccount(input.ctx, addr1)
	require.True(t, getAcc.GetCoins().IsEqual(coins))
	require.Len(t, events, 2)
}

// Test that all hardcoded VM Path are correct.
// If something goes wrong, check the DataSource logs for requested Path and fix.
func TestKeeper_Path(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// Create account
	baseAmount := sdk.NewInt(1000)
	accCoins := sdk.NewCoins(
		sdk.NewCoin("dfi", baseAmount),
		sdk.NewCoin("eth", baseAmount),
		sdk.NewCoin("btc", baseAmount),
		sdk.NewCoin("usdt", baseAmount),
	)

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	acc1.SetCoins(accCoins)
	input.ak.SetAccount(input.ctx, acc1)

	// Init genesis and start DS
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// Launch DVM compiler and runtime
	client, compiler, vm := launchDocker(input.dsPort, t)
	defer input.vk.CloseConnections()
	defer stopDocker(t, client, compiler)
	defer stopDocker(t, client, vm)

	require.NoError(t, waitStarted(client, compiler.ID, 5*time.Second), "DVM compiler: start")
	require.NoError(t, waitStarted(client, vm.ID, 5*time.Second), "DVM runtime: start")
	require.NoError(t, waitReachable(*vmCompiler, 5*time.Second), "DVM compiler: wait to be reachable")
	require.NoError(t, waitReachable(*vmAddress, 5*time.Second), "DVM runtime: wait to be reachable")

	// Check middleware path: Block
	testID := "Middleware Block"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Block;
    			fun main() {
        			let _ = Block::get_current_block_height();
    			}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check middleware path: Time
	testID = "middleware Time"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x0::Time;
			    fun main() {
        			let _ = Time::now();
    			}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: DFI
	testID = "VMAuth DFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x0::Account;
				use 0x0::DFI;
				fun main() {
					let _ = Account::balance<DFI::T>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: ETH
	testID = "VMAuth ETH"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x0::Account;
				use 0x0::Coins;
				fun main() {
					let _ = Account::balance<Coins::ETH>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: USDT
	testID = "VMAuth USDT"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x0::Account;
				use 0x0::Coins;
				fun main() {
					let _ = Account::balance<Coins::USDT>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: BTC
	testID = "VMAuth BTC"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x0::Account;
				use 0x0::Coins;
				fun main() {
					let _ = Account::balance<Coins::BTC>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: DFI
	testID = "CurrencyInfo DFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Dfinance;
				use 0x0::DFI;
				fun main() {
					let _ = Dfinance::denom<DFI::T>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: ETH
	testID = "CurrencyInfo ETH"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Dfinance;
				use 0x0::Coins;
				fun main() {
					let _ = Dfinance::denom<Coins::ETH>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: USDT
	testID = "CurrencyInfo USDT"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Dfinance;
				use 0x0::Coins;
				fun main() {
					let _ = Dfinance::denom<Coins::USDT>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: BTC
	testID = "CurrencyInfo BTC"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Dfinance;
				use 0x0::Coins;
				fun main() {
					let _ = Dfinance::denom<Coins::BTC>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check VMAuth module path: Account eventHandler
	testID = "VMAuth eventHandler"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x0::Event;
				fun main() {
					let event_handle = Event::new_event_handle<u64>();
					Event::emit_event(&mut event_handle, 1);
					Event::destroy_handle(event_handle);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Check VMAuth module path: Account resource
	testID = "VMAuth accResource"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := 	`
			script {
				use 0x0::Account;
				use 0x0::DFI;
				fun main() {
					let dfi = Account::withdraw_from_sender<DFI::T>(1);
					Account::deposit_to_sender<DFI::T>(dfi);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}

	// Create module and use it in script (doesn't check VM path)
	testID = "Account module"
	{
		t.Logf("%s: module compile", testID)
		moduleSrc := `
			module Dummy {
			    public fun one_u64(): u64 {
					1
			    }
			}
		`
		moduleBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    moduleSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Module,
		})
		require.NoErrorf(t, err, "%s: module compile error", testID)

		t.Logf("%s: module deploy", testID)
		moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode)
		require.NoErrorf(t, moduleMsg.ValidateBasic(), "%s: module deploy message validation failed", testID)
		ctx, writeCtx := input.ctx.CacheContext()
		require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "%s: module deploy error", testID)

		t.Logf("%s: checking module events", testID)
		checkNoErrors(ctx.EventManager().Events(), t)
		writeCtx()

		t.Logf("%s: script compile", testID)
		scriptSrcFmt := `
			script {
				use %s::Dummy;
    			fun main() {
       			let _ = Dummy::one_u64();
    			}
			}
		`
		scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
		scriptBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
			Type:    vm_grpc.ContractType_Script,
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoErrors(input.ctx.EventManager().Events(), t)
	}
}
