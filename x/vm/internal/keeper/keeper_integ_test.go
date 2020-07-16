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

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnodeConfig "github.com/dfinance/dnode/cmd/config"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle"
	compilerClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const sendScript = `
script {
	use 0x1::Account;
	use 0x1::Coins;
	use 0x1::DFI;
	use 0x1::Dfinance;
	use 0x1::Compare;
	
	fun main(account: &signer, recipient: address, dfi_amount: u128, eth_amount: u128, btc_amount: u128, usdt_amount: u128) {
		Account::pay_from_sender<DFI::T>(account, recipient, dfi_amount);
		Account::pay_from_sender<Coins::ETH>(account, recipient, eth_amount);
		Account::pay_from_sender<Coins::BTC>(account, recipient, btc_amount);
		Account::pay_from_sender<Coins::USDT>(account, recipient, usdt_amount);

		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<DFI::T>(), &b"dfi") == 0, 1);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::ETH>(), &b"eth") == 0, 2);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::BTC>(), &b"btc") == 0, 3);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::USDT>(), &b"usdt") == 0, 4);

		assert(Dfinance::decimals<DFI::T>() == 18, 5);
		assert(Dfinance::decimals<Coins::ETH>() == 18, 6);
		assert(Dfinance::decimals<Coins::BTC>() == 8, 7);
		assert(Dfinance::decimals<Coins::USDT>() == 6, 8);
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
	use 0x1::Event;
	use {{sender}}::Math;
	
	fun main(_account: &signer, a: u64, b: u64) {
		let c = Math::add(a, b);
		Event::emit<u64>(c);
	}
}
`

const oraclePriceScript = `
script {
	use 0x1::Event;
	use 0x1::Oracle;
	use 0x1::Coins;

	fun main(_account: &signer) {
		let price = Oracle::get_price<Coins::ETH, Coins::USDT>();
		Event::emit<u64>(price);
	}
}
`

const errorScript = `
script {
	use 0x1::Account;
	use 0x1::DFI;
	use 0x1::Coins;
	use 0x1::Event;

	fun main(account: &signer, c: u64) {
		let a = Account::withdraw_from_sender<DFI::T>(account, 523);
		let b = Account::withdraw_from_sender<Coins::BTC>(account, 1);
	
		Event::emit<u64>(10);
	
		assert(c == 1000, 122);
		Account::deposit_to_sender(account, a);
		Account::deposit_to_sender(account, b);
	}
}
`

const argsScript = `
script {
	use 0x1::Vector;

	fun main(account: &signer, arg_u8: u8, arg_u64: u64, arg_u128: u128, arg_addr: address, arg_bool_true: bool, arg_bool_false: bool, arg_vector: vector<u8>) {
        assert(arg_u8 == 128, 10);
        assert(arg_u64 == 1000000, 11);
        assert(arg_u128 == 100000000000000000000000000000, 12);
        
        assert(0x1::Signer::address_of(account) == arg_addr, 20);

        assert(arg_bool_true == true, 30);
        assert(arg_bool_false == false, 31);
        
        assert(Vector::length<u8>(&mut arg_vector) == 2, 40);
        assert(Vector::pop_back<u8>(&mut arg_vector) == 1, 41);
        assert(Vector::pop_back<u8>(&mut arg_vector) == 0, 42);
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
func checkEventSubStatus(events sdk.Events, subStatus uint64, t *testing.T) {
	found := false

	for _, event := range events {
		if event.Type == types.EventTypeContractStatus {
			// find error
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeErrSubStatus {
					require.EqualValues(t, attr.Value, []byte(strconv.FormatUint(subStatus, 10)), "wrong value for sub status")

					found = true
				}
			}
		}
	}

	require.True(t, found, "sub status not found")
}

// check script doesn't contains errors.
func checkNoEventErrors(events sdk.Events, t *testing.T) {
	for _, event := range events {
		if event.Type == types.EventTypeContractStatus {
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeStatus {
					if string(attr.Value) == types.AttributeValueStatusDiscard {
						printEvent(event, t)
						t.Fatalf("should not contains error event")
					}

					if string(attr.Value) == types.AttributeValueStatusError {
						printEvent(event, t)
						t.Fatalf("should not contains error event")
					}
				}
			}
		}
	}
}

// check that eventsA contains every event of eventsB
func checkEventsContainsEvery(t *testing.T, eventsA, eventsB sdk.Events) {
	require.GreaterOrEqual(t, len(eventsA), len(eventsB), "events length mismatch: %d / %d", len(eventsA), len(eventsB))
	for i, event := range eventsB {
		require.Contains(t, eventsA, event, "doesn't contain event[%d]", i)
	}
}

// creates "keep" without an error events.
func newKeepEvents() sdk.Events {
	return types.NewContractEvents(&vm_grpc.VMExecuteResponse{Status: vm_grpc.ContractStatus_Keep})
}

// Test transfer of dfi between two accounts in dfi.
func TestKeeper_DeployContractTransfer(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// launch docker
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

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

	t.Logf("Compile send script")
	bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    sendScript,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for send script: %v", err)

	// execute contract.
	var args []types.ScriptArg
	{
		arg, err := compilerClient.NewAddressScriptArg(addr2.String())
		require.NoError(t, err)
		args = append(args, arg)
	}
	for _, d := range denoms {
		arg, err := compilerClient.NewU128ScriptArg(toSend[d].String())
		require.NoError(t, err)
		args = append(args, arg)
	}

	t.Logf("Execute send script")
	msgScript := types.NewMsgExecuteScript(addr1, bytecode, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	checkEventsContainsEvery(t, events, newKeepEvents())

	checkNoEventErrors(events, t)

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

	input := newTestInput(false)

	// launch docker
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)

	input.ak.SetAccount(input.ctx, acc1)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	bytecodeModule, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    mathModule,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for math module: %v", err)

	msg := types.NewMsgDeployModule(addr1, bytecodeModule)
	err = msg.ValidateBasic()
	require.NoErrorf(t, err, "can't validate err: %v", err)

	ctx, writeCtx := input.ctx.CacheContext()
	err = input.vk.DeployContract(ctx, msg)
	require.NoErrorf(t, err, "can't deploy contract: %v", err)

	events := ctx.EventManager().Events()
	checkNoEventErrors(events, t)

	writeCtx()

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    strings.Replace(mathScript, "{{sender}}", addr1.String(), 1),
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't compiler script for math module: %v", err)

	var args []types.ScriptArg
	{
		arg, err := compilerClient.NewU64ScriptArg("10")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewU64ScriptArg("100")
		require.NoError(t, err)
		args = append(args, arg)
	}

	ctx, _ = input.ctx.CacheContext()
	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, args)
	err = input.vk.ExecuteScript(ctx, msgScript)
	require.NoError(t, err)

	events = ctx.EventManager().Events()
	checkNoEventErrors(events, t)

	checkEventsContainsEvery(t, events, newKeepEvents())
	vmEvent := events[2]
	require.Equal(t, vmEvent.Type, types.EventTypeMoveEvent, "script after execution doesn't contain event with amount")
	require.Len(t, vmEvent.Attributes, 4)
	// sender
	{
		attrIdx := 0
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventSender)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifySenderAddress(addr1))
	}
	// source
	{
		attrIdx := 1
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventSource)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.GetEventSourceAttribute(nil))
	}
	// type
	{
		attrIdx := 2
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventType)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), &vm_grpc.LcsTag{TypeTag: vm_grpc.LcsType_LcsU64}))
	}
	// data
	{
		attrIdx := 3
		uintBz := make([]byte, 8)
		binary.LittleEndian.PutUint64(uintBz, uint64(110))
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventData)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, hex.EncodeToString(uintBz))
	}
}

// Test oracle price return.
func TestKeeper_ScriptOracle(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// launch docker
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// create accounts.
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)

	input.ak.SetAccount(input.ctx, acc1)

	assetCode := dnTypes.AssetCode("eth_usdt")
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

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    oraclePriceScript,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for oracle script: %v", err)

	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, nil)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	checkNoEventErrors(events, t)

	checkEventsContainsEvery(t, events, newKeepEvents())
	require.Len(t, events, 3)
	vmEvent := events[2]
	require.Len(t, vmEvent.Attributes, 4)
	// sender
	{
		attrIdx := 0
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventSender)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifySenderAddress(addr1))
	}
	// source
	{
		attrIdx := 1
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventSource)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.GetEventSourceAttribute(nil))
	}
	// type
	{
		attrIdx := 2
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventType)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), &vm_grpc.LcsTag{TypeTag: vm_grpc.LcsType_LcsU64}))
	}
	// data
	{
		attrIdx := 3
		uintBz := make([]byte, 8)
		binary.LittleEndian.PutUint64(uintBz, 100)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventData)
		require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, hex.EncodeToString(uintBz))
	}
}

// Test oracle price return.
func TestKeeper_ErrorScript(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// launch docker
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

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

	bytecodeScript, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    errorScript,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for error script: %v", err)

	var args []types.ScriptArg
	{
		arg, err := compilerClient.NewU64ScriptArg(strconv.FormatUint(10, 10))
		require.NoError(t, err)
		args = append(args, arg)
	}

	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	checkEventsContainsEvery(t, events, newKeepEvents())
	for _, e := range events {
		printEvent(e, t)
	}
	checkEventSubStatus(events, 122, t)

	// first of all - check balance
	// then check that error still there
	// then check that no events there only error and keep status
	getAcc := input.ak.GetAccount(input.ctx, addr1)
	require.True(t, getAcc.GetCoins().IsEqual(coins))
	require.Len(t, events, 3)
}

func TestKeeper_AllArgsTypes(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	accCoins := sdk.NewCoins(sdk.NewCoin("dfi", sdk.NewInt(1000)))
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

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Compile script
	bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    argsScript,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")

	// Add all args and execute
	var args []types.ScriptArg
	{
		arg, err := compilerClient.NewU8ScriptArg("128")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewU64ScriptArg("1000000")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewU128ScriptArg("100000000000000000000000000000")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewAddressScriptArg(addr1.String())
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewBoolScriptArg("true")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewBoolScriptArg("false")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := compilerClient.NewVectorScriptArg("0x0001")
		require.NoError(t, err)
		args = append(args, arg)
	}

	scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, args)
	require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "script execute error")

	checkNoEventErrors(input.ctx.EventManager().Events(), t)
}

// Test that all hardcoded VM Path are correct.
// If something goes wrong, check the DataSource logs for requested Path and fix.
func TestKeeper_Path(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

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

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Check middleware path: Block
	testID := "Middleware Block"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Block;

    			fun main() {
        			let _ = Block::get_current_block_height();
    			}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check middleware path: Time
	testID = "Middleware Time"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Time;

			    fun main() {
        			let _ = Time::now();
    			}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: DFI
	testID = "VMAuth DFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Account;
				use 0x1::DFI;

				fun main(account: &signer) {
					let _ = Account::balance<DFI::T>(account);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: ETH
	testID = "VMAuth ETH"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Account;
				use 0x1::Coins;

				fun main(account: &signer) {
					let _ = Account::balance<Coins::ETH>(account);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: USDT
	testID = "VMAuth USDT"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Account;
				use 0x1::Coins;

				fun main(account: &signer) {
					let _ = Account::balance<Coins::USDT>(account);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: BTC
	testID = "VMAuth BTC"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Account;
				use 0x1::Coins;

				fun main(account: &signer) {
					let _ = Account::balance<Coins::BTC>(account);
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: DFI
	testID = "CurrencyInfo DFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Dfinance;
				use 0x1::DFI;

				fun main() {
					let _ = Dfinance::denom<DFI::T>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: ETH
	testID = "CurrencyInfo ETH"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Dfinance;
				use 0x1::Coins;

				fun main() {
					let _ = Dfinance::denom<Coins::ETH>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: USDT
	testID = "CurrencyInfo USDT"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Dfinance;
				use 0x1::Coins;

				fun main() {
					let _ = Dfinance::denom<Coins::USDT>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: BTC
	testID = "CurrencyInfo BTC"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Dfinance;
				use 0x1::Coins;

				fun main() {
					let _ = Dfinance::denom<Coins::BTC>();
				}
			}
		`
		bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
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
		moduleBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    moduleSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: module compile error", testID)

		t.Logf("%s: module deploy", testID)
		moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode)
		require.NoErrorf(t, moduleMsg.ValidateBasic(), "%s: module deploy message validation failed", testID)
		ctx, writeCtx := input.ctx.CacheContext()
		require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "%s: module deploy error", testID)

		t.Logf("%s: checking module events", testID)
		checkNoEventErrors(ctx.EventManager().Events(), t)
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
		scriptBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
			Text:    scriptSrc,
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}
}

// VM Event.EventType string serialization test.
func Test_EventTypeSerialization(t *testing.T) {
	const moduleSrc = `
		module Foo {
		    struct FooEvent<T, VT> {
		        field_T:  T,
		        field_VT: VT
		    }
		
		    public fun NewFooEvent<T, VT>(arg_T: T, arg_VT: VT): FooEvent<T, VT> {
		        let fooEvent = FooEvent<T, VT> {
		            field_T:  arg_T,
		            field_VT: arg_VT
		        };
				
				0x1::Event::emit<bool>(true);
		
		        fooEvent
		    }
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::Foo;
			
			fun main(_account: &signer) {
				// Event with single tag
				0x1::Event::emit<u8>(128);
				
				// Event with single vector
				0x1::Event::emit<vector<u8>>(x"0102");
				
				// Two events:
				//   1. Module: single tag
				//   2. Script: generic struct with tag, vector
				let fooEvent = Foo::NewFooEvent<u64, vector<u8>>(1000, x"0102");
				0x1::Event::emit<Foo::FooEvent<u64, vector<u8>>>(fooEvent);
    		}
		}
	`

	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	input.ak.SetAccount(input.ctx, acc1)

	// Init genesis and start DS
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Compile, publish module
	moduleBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    moduleSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    scriptSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")

	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode, nil)
	resp, err := input.vk.ExecuteScriptNoProcessing(input.ctx, scriptMsg)
	require.NoErrorf(t, err, "script execute error")

	t.Logf("checking script events")
	checkNoEventErrors(input.ctx.EventManager().Events(), t)

	for idx, event := range resp.Events {
		t.Logf("VM Event #%d", idx)
		t.Log(types.VMEventToString(event))

		t.Logf("Cosmos Event #%d", idx)
		cosmosEvent := types.NewMoveEvent(sdk.NewInfiniteGasMeter(), event)
		printEvent(cosmosEvent, t)
	}
}

// VM Event.EventType string serialization test with gas charged check.
func Test_EventTypeSerializationGas(t *testing.T) {
	const moduleSrc = `
		module GasEvent {
			struct A {
				value: u64
			}
		
			struct B<T> {
				value: T
			}
		
			struct C<T> {
				value: T
			}
		
			struct D<T> {
				value: T
			}
		
			public fun test() {
				let a = A {
					value: 10
				};
		
				let b = B<A> {
					value: a
				};
		
				let c = C<B<A>> {
					value: b
				};
		
				let d = D<C<B<A>>> {
					value: c
				};
		
				0x1::Event::emit<D<C<B<A>>>>(d);
			}
		
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::GasEvent;

			fun main() {
				GasEvent::test();
			}
		}
	`

	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	input.ak.SetAccount(input.ctx, acc1)

	// Init genesis and start DS
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Compile, publish module
	moduleBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    moduleSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    scriptSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")

	gasMeter := sdk.NewGasMeter(100000)
	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode, nil)
	scriptErr := input.vk.ExecuteScript(input.ctx.WithGasMeter(gasMeter), scriptMsg)
	require.NoError(t, scriptErr, "script execute error")

	// Calculate min / max gasConsumed: script has 4 depth level
	expectedMinChargedGas := uint64(0)
	for i := 1; i <= 4-types.EventTypeNoGasLevels; i++ {
		expectedMinChargedGas += uint64(i) * types.EventTypeProcessingGas
	}
	expectedMaxChargedGas := expectedMinChargedGas + types.EventTypeProcessingGas

	t.Logf("Consumed gas: %d", gasMeter.GasConsumed())
	t.Logf("Expected min/max gas: %d / %d", expectedMinChargedGas, expectedMaxChargedGas)
	require.GreaterOrEqual(t, gasMeter.GasConsumed(), expectedMinChargedGas)
	require.LessOrEqual(t, gasMeter.GasConsumed(), expectedMaxChargedGas)
}

// VM Event.EventType string serialization test with out of gas.
func Test_EventTypeSerializationOutOfGas(t *testing.T) {
	const moduleSrc = `
		module OutOfGasEvent {
			struct A {
				value: u64
			}
		
			struct B<T> {
				value: T
			}
		
			struct C<T> {
				value: T
			}
		
			struct Z<T> {
				value: T
			}
		
			struct V<T> {
				value: T
			}
		
			struct M<T> {
				value: T
			}
		
			public fun test() {
				let a = A {
					value: 10
				};
		
				let b = B<A> {
					value: a
				};
		
				let c = C<B<A>> {
					value: b
				};
		
				let z = Z<C<B<A>>> {
					value: c
				};
		
				let v = V<Z<C<B<A>>>> {
					value: z
				};
		
				let m = M<V<Z<C<B<A>>>>> {
					value: v
				};
		
				0x1::Event::emit<M<V<Z<C<B<A>>>>>>(m);
			}
		
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::OutOfGasEvent;

			fun main() {
				OutOfGasEvent::test();
			}
		}
	`

	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	input.ak.SetAccount(input.ctx, acc1)

	// Init genesis and start DS
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Compile, publish module
	moduleBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    moduleSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.SourceFile{
		Text:    scriptSrc,
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")

	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode, nil)

	require.PanicsWithValue(t, sdk.ErrorOutOfGas{"event type processing"}, func() {
		input.vk.ExecuteScript(input.ctx.WithGasMeter(sdk.NewGasMeter(100000)), scriptMsg)
	})
}
