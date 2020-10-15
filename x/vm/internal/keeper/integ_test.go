// +build integ

package keeper

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/compiler_grpc"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnodeConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const sendScript = `
script {
	use 0x1::Account;
	use 0x1::Coins;
	use 0x1::XFI;
	use 0x1::Dfinance;
	use 0x1::Compare;
	
	fun main(account: &signer, recipient: address, xfi_amount: u128, eth_amount: u128, btc_amount: u128, usdt_amount: u128) {
		Account::pay_from_sender<XFI::T>(account, recipient, xfi_amount);
		Account::pay_from_sender<Coins::ETH>(account, recipient, eth_amount);
		Account::pay_from_sender<Coins::BTC>(account, recipient, btc_amount);
		Account::pay_from_sender<Coins::USDT>(account, recipient, usdt_amount);

		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<XFI::T>(), &b"xfi") == 0, 1);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::ETH>(), &b"eth") == 0, 2);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::BTC>(), &b"btc") == 0, 3);
		assert(Compare::cmp_lcs_bytes(&Dfinance::denom<Coins::USDT>(), &b"usdt") == 0, 4);

		assert(Dfinance::decimals<XFI::T>() == 18, 5);
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
	
	fun main(account: &signer, a: u64, b: u64) {
		let c = Math::add(a, b);
		Event::emit<u64>(account, c);
	}
}
`

const mathDoubleModule = `
module DblMathAdd {
    public fun add(a: u64, b: u64): u64 {
		a + b
    }
}

module DblMathSub {
	public fun sub(a: u64, b: u64): u64 {
		a - b
    }
}
`

const mathDoubleScript = `
script {
	use 0x1::Event;
	use {{sender}}::DblMathAdd;
	use {{sender}}::DblMathSub;
	
	fun main(account: &signer, a: u64, b: u64, c: u64) {
		let ab = DblMathAdd::add(a, b);
		let res = DblMathSub::sub(ab, c);
		Event::emit<u64>(account, res);
	}
}

script {
	use 0x1::Event;
	use {{sender}}::DblMathAdd;
	use {{sender}}::DblMathSub;
	
	fun main(account: &signer, a: u64, b: u64, c: u64) {
		let ab = DblMathSub::sub(a, b);
		let res = DblMathAdd::add(ab, c);
		Event::emit<u64>(account, res);
	}
}
`

const oraclePriceScript = `
script {
	use 0x1::Event;
	use 0x1::Coins;

	fun main(account: &signer) {
		let price = Coins::get_price<Coins::ETH, Coins::USDT>();
		Event::emit<u128>(account, price);
	}
}
`
const oracleReverseAssetPriceScript = `
script {
	use 0x1::Event;
	use 0x1::Coins;

	fun main(account: &signer) {
		let price = Coins::get_price<Coins::USDT, Coins::ETH>();
		Event::emit<u128>(account, price);
	}
}
`

const errorScript = `
script {
	use 0x1::Account;
	use 0x1::XFI;
	use 0x1::Coins;
	use 0x1::Event;

	fun main(account: &signer, c: u64) {
		let a = Account::withdraw_from_sender<XFI::T>(account, 523);
		let b = Account::withdraw_from_sender<Coins::BTC>(account, 1);
	
		Event::emit<u64>(account, 10);
	
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
		t.Logf("  %s = %s\n", attr.Key, attr.Value)
		if string(attr.Key) == types.AttributeErrMajorStatus {
			errMsg := types.GetStrCode(string(attr.Value))
			t.Logf("  %s description: %s\n", attr.Key, errMsg)
		}
	}
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
				}
			}
		}
	}
}

// check events contains errors.
func checkEventErrors(events sdk.Events) bool {
	for _, event := range events {
		if event.Type == types.EventTypeContractStatus {
			for _, attr := range event.Attributes {
				if string(attr.Key) == types.AttributeStatus {
					if string(attr.Value) == types.AttributeValueStatusDiscard {
						return true
					}
				}
			}
		}
	}

	return false
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
	return types.NewContractEvents(&vm_grpc.VMExecuteResponse{
		Status: &vm_grpc.VMStatus{},
	})
}

// Test transfer of xfi between two accounts in xfi.
func TestVMKeeper_DeployContractTransfer(t *testing.T) {
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
		sdk.NewCoin("xfi", baseAmount),
		sdk.NewCoin("eth", baseAmount),
		sdk.NewCoin("btc", baseAmount),
		sdk.NewCoin("usdt", baseAmount),
	)

	denoms := make([]string, 4)
	denoms[0] = "xfi"
	denoms[1] = "eth"
	denoms[2] = "btc"
	denoms[3] = "usdt"

	toSend := make(map[string]sdk.Int, 4)

	for i := 0; i < len(denoms); i++ {
		toSend[denoms[i]] = sdk.NewInt(100 - int64(i)*10)
	}

	_ = acc1.SetCoins(putCoins)

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
	bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: sendScript,
				Name: "SendScript",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for send script: %v", err)
	require.Len(t, bytecode, 1)

	// execute contract.
	var args []types.ScriptArg
	{
		arg, err := vm_client.NewAddressScriptArg(addr2.String())
		require.NoError(t, err)
		args = append(args, arg)
	}
	for _, d := range denoms {
		arg, err := vm_client.NewU128ScriptArg(toSend[d].String())
		require.NoError(t, err)
		args = append(args, arg)
	}

	t.Logf("Execute send script")
	msgScript := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, args)
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
func TestVMKeeper_DeployModule(t *testing.T) {
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

	bytecodeModule, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: mathModule,
				Name: "MathModule",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for math module: %v", err)
	require.Len(t, bytecodeModule, 1)

	msg := types.NewMsgDeployModule(addr1, bytecodeModule[0].ByteCode)
	err = msg.ValidateBasic()
	require.NoErrorf(t, err, "can't validate err: %v", err)

	ctx, writeCtx := input.ctx.CacheContext()
	err = input.vk.DeployContract(ctx, msg)
	require.NoErrorf(t, err, "can't deploy contract: %v", err)

	events := ctx.EventManager().Events()
	checkNoEventErrors(events, t)

	writeCtx()

	bytecodeScript, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: strings.Replace(mathScript, "{{sender}}", addr1.String(), 1),
				Name: "MathScript",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't compiler script for math module: %v", err)
	require.Len(t, bytecodeScript, 1)

	var args []types.ScriptArg
	{
		arg, err := vm_client.NewU64ScriptArg("10")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewU64ScriptArg("100")
		require.NoError(t, err)
		args = append(args, arg)
	}

	ctx, _ = input.ctx.CacheContext()
	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript[0].ByteCode, args)
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

// Test deploy the same module twice.
func TestVMKeeper_DeployModuleTwice(t *testing.T) {
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

	checkModuleCompiled := func(msg string, srcCode string) []vm_client.CompiledItem {
		byteCode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: srcCode,
					Name: "checkModuleCompiled",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoError(t, err, "%s: can't get code for module", msg)
		return byteCode
	}

	checkMsgCreate := func(msg string, byteCode []byte) types.MsgDeployModule {
		deployMsg := types.NewMsgDeployModule(addr1, byteCode)
		require.NoError(t, deployMsg.ValidateBasic(), "%s: can't validate err: %v", msg)
		return deployMsg
	}

	checkDeployOK := func(msg string, byteCode []byte) {
		deployMsg := checkMsgCreate(msg, byteCode)

		ctx, writeCtx := input.ctx.CacheContext()
		err := input.vk.DeployContract(ctx, deployMsg)
		require.NoError(t, err, "%s: can't deploy contract", msg)

		events := ctx.EventManager().Events()
		checkNoEventErrors(events, t)

		writeCtx()
	}

	checkDeployFailed := func(msg string, byteCode []byte) {
		deployMsg := checkMsgCreate(msg, byteCode)

		ctx, _ := input.ctx.CacheContext()
		err := input.vk.DeployContract(ctx, deployMsg)
		require.NoError(t, err, "%s: can't deploy contract", msg)

		events := ctx.EventManager().Events()
		checkEventsContainsEvery(t, events, sdk.Events{
			sdk.NewEvent(
				types.EventTypeContractStatus,
				sdk.NewAttribute(types.AttributeStatus, types.AttributeValueStatusDiscard),
				sdk.NewAttribute(types.AttributeErrMajorStatus, "1095"),
				sdk.NewAttribute(types.AttributeErrSubStatus, "0"),
			),
		})
	}

	checkScriptCompiled := func(msg, srcCode string) []vm_client.CompiledItem {
		srcCode = strings.Replace(srcCode, "{{sender}}", addr1.String(), -1)
		byteCode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: srcCode,
					Name: "MathScript",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoError(t, err, "%s: can't compiler script", msg)
		return byteCode
	}

	checkScriptNotCompiled := func(msg, srcCode string) {
		srcCode = strings.Replace(srcCode, "{{sender}}", addr1.String(), -1)
		_, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: srcCode,
					Name: "MathScript",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.Error(t, err, msg)
	}

	_ = checkScriptNotCompiled

	checkScriptExecuteOK := func(msg string, byteCode []byte, args []types.ScriptArg) {
		ctx, _ := input.ctx.CacheContext()
		executeMsg := types.NewMsgExecuteScript(addr1, byteCode, args)
		err := input.vk.ExecuteScript(ctx, executeMsg)
		require.NoError(t, err, "%s: can't execute script", msg)
	}

	var mathScriptArgs []types.ScriptArg
	{
		arg, err := vm_client.NewU64ScriptArg("10")
		require.NoError(t, err)
		mathScriptArgs = append(mathScriptArgs, arg)
	}
	{
		arg, err := vm_client.NewU64ScriptArg("100")
		require.NoError(t, err)
		mathScriptArgs = append(mathScriptArgs, arg)
	}

	// test case 1
	{
		moduleByteCode := checkModuleCompiled("TestCase 1", mathModule)
		require.Len(t, moduleByteCode, 1)
		checkDeployOK("TestCase 1: 1st deploy", moduleByteCode[0].ByteCode)
		checkDeployFailed("TestCase 1: 2nd deploy", moduleByteCode[0].ByteCode)

		scriptByteCode := checkScriptCompiled("TestCase 1", mathScript)
		require.Len(t, scriptByteCode, 1)
		checkScriptExecuteOK("TestCase 1", scriptByteCode[0].ByteCode, mathScriptArgs)
	}

	// test case 2: module with "address 0x... {}" prefix
	{
		moduleSrcCode := fmt.Sprintf("address %s {\n%s\n}", addr1, strings.Replace(mathModule, "Math", "Math2", 1))
		scriptSrcCode := strings.Replace(mathScript, "Math", "Math2", -1)

		moduleByteCode := checkModuleCompiled("TestCase 2", moduleSrcCode)
		require.Len(t, moduleByteCode, 1)
		checkDeployOK("TestCase 2: 1st deploy", moduleByteCode[0].ByteCode)
		checkDeployFailed("TestCase 2: 2nd deploy", moduleByteCode[0].ByteCode)

		scriptByteCode := checkScriptCompiled("TestCase 2", scriptSrcCode)
		require.Len(t, scriptByteCode, 1)
		checkScriptExecuteOK("TestCase 2", scriptByteCode[0].ByteCode, mathScriptArgs)
	}

	//var dblMathScriptArgs []types.ScriptArg
	//{
	//	arg, err := vm_client.NewU64ScriptArg("10")
	//	require.NoError(t, err)
	//	mathScriptArgs = append(mathScriptArgs, arg)
	//}
	//{
	//	arg, err := vm_client.NewU64ScriptArg("100")
	//	require.NoError(t, err)
	//	mathScriptArgs = append(mathScriptArgs, arg)
	//}
	//{
	//	arg, err := vm_client.NewU64ScriptArg("20")
	//	require.NoError(t, err)
	//	mathScriptArgs = append(mathScriptArgs, arg)
	//}

	// test case 3: two module in one
	{
		moduleByteCode := checkModuleCompiled("TestCase 3", mathDoubleModule)
		require.Len(t, moduleByteCode, 2)
		checkDeployOK("TestCase 3: 1st deploy", moduleByteCode[0].ByteCode)
		checkDeployOK("TestCase 3: 2nd deploy", moduleByteCode[1].ByteCode)

		scriptByteCode := checkScriptCompiled("TestCase 3", mathDoubleScript)
		require.Len(t, scriptByteCode, 2)
		checkScriptExecuteOK("TestCase 3: execute", scriptByteCode[0].ByteCode, mathScriptArgs)
	}

	// test case 4: two module in one, module srcCode with "address 0x... {}" prefix
	{
		moduleSrcCode := fmt.Sprintf("address %s {\n%s\n}", addr1, strings.Replace(mathDoubleModule, "DblMath", "DblMath2", 2))
		scriptSrcCode := strings.Replace(mathDoubleScript, "DblMath", "DblMath2", -1)

		moduleByteCode := checkModuleCompiled("TestCase 4", moduleSrcCode)
		require.Len(t, moduleByteCode, 2)
		checkDeployOK("TestCase 4: 1st deploy", moduleByteCode[0].ByteCode)
		checkDeployOK("TestCase 4: 2nd deploy", moduleByteCode[1].ByteCode)

		scriptByteCode := checkScriptCompiled("TestCase 4", scriptSrcCode)
		require.Len(t, scriptByteCode, 2)
		checkScriptExecuteOK("TestCase 4: execute", scriptByteCode[0].ByteCode, mathScriptArgs)
	}
}

// Test oracle price return.
func TestVMKeeper_ScriptOracle(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)
	price := sdk.NewInt(10).Mul(sdk.NewInt(int64(math.Pow10(oracle.PricePrecision))))

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
	_, _ = input.ok.SetPrice(input.ctx, addr1, assetCode, price, price, time.Now())
	_ = input.ok.SetCurrentPrices(input.ctx)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	// compile direct asset
	{
		bytecodeScript, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: oraclePriceScript,
					Name: "OraclePriceScript",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "can't get code for oracle direct asset script: %v", err)
		require.Len(t, bytecodeScript, 1)

		msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript[0].ByteCode, nil)
		err = input.vk.ExecuteScript(input.ctx, msgScript)
		require.NoError(t, err)

		events := input.ctx.EventManager().Events()
		checkNoEventErrors(events, t)

		checkEventsContainsEvery(t, events, newKeepEvents())
		require.Len(t, events, 5)
		vmEvent := events[4]
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
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), &vm_grpc.LcsTag{TypeTag: vm_grpc.LcsType_LcsU128}))
		}
		// data
		{
			attrIdx := 3
			price := sdk.NewIntFromBigInt(price.BigInt())
			priceBz := helpers.BigToBytes(price, 16) // u128
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventData)
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, hex.EncodeToString(priceBz))
		}
	}

	// compile reverse asset
	{
		events := input.ctx.EventManager().Events()
		bytecodeReverseScript, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: oracleReverseAssetPriceScript,
					Name: "oracleReverseAssetPriceScript",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "can't get code for oracle reverse asset script: %v", err)

		msgScript := types.NewMsgExecuteScript(addr1, bytecodeReverseScript[0].ByteCode, nil)
		err = input.vk.ExecuteScript(input.ctx, msgScript)
		require.NoError(t, err)

		events = input.ctx.EventManager().Events()
		checkNoEventErrors(events, t)

		checkEventsContainsEvery(t, events, newKeepEvents())
		require.Len(t, events, 8)
		vmEvent := events[7]
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
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, types.StringifyEventTypePanic(sdk.NewInfiniteGasMeter(), &vm_grpc.LcsTag{TypeTag: vm_grpc.LcsType_LcsU128}))
		}
		// data
		{
			attrIdx := 3
			// price calculation for price 10.0000 in float: 1/10.000 => 0.0100 (10/100 => 0.0100)
			// in test for the Int with 8 digit precision
			// 10 0000 0000 reverse price is 1000 0000,
			// so 10 0000 0000 / 100 => 1000 0000
			clcPrice := price.Quo(sdk.NewInt(100))
			priceBz := helpers.BigToBytes(clcPrice, 16) // u128
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Key, types.AttributeVmEventData)
			require.EqualValues(t, vmEvent.Attributes[attrIdx].Value, hex.EncodeToString(priceBz))
		}
	}
}

// Test oracle price return.
func TestVMKeeper_ErrorScript(t *testing.T) {
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
		sdk.NewCoin("xfi", sdk.NewInt(1000000000000000)),
		sdk.NewCoin("btc", sdk.NewInt(1)),
	)

	_ = acc1.SetCoins(coins)
	input.ak.SetAccount(input.ctx, acc1)

	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(2 * time.Second)

	bytecodeScript, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: errorScript,
				Name: "errorScript",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "can't get code for error script: %v", err)
	require.Len(t, bytecodeScript, 1)

	var args []types.ScriptArg
	{
		arg, err := vm_client.NewU64ScriptArg(strconv.FormatUint(10, 10))
		require.NoError(t, err)
		args = append(args, arg)
	}

	msgScript := types.NewMsgExecuteScript(addr1, bytecodeScript[0].ByteCode, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	require.True(t, checkEventErrors(events))

	// first of all - check balance
	// then check that error still there
	// then check that no events there only error and keep status
	getAcc := input.ak.GetAccount(input.ctx, addr1)
	require.True(t, getAcc.GetCoins().IsEqual(coins))
	require.Len(t, events, 2)
}

func TestVMKeeper_AllArgsTypes(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	accCoins := sdk.NewCoins(sdk.NewCoin("xfi", sdk.NewInt(1000)))
	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	_ = acc1.SetCoins(accCoins)
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
	bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: argsScript,
				Name: "argsScript",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")
	require.Len(t, bytecode, 1)

	// Add all args and execute
	var args []types.ScriptArg
	{
		arg, err := vm_client.NewU8ScriptArg("128")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewU64ScriptArg("1000000")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewU128ScriptArg("100000000000000000000000000000")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewAddressScriptArg(addr1.String())
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewBoolScriptArg("true")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewBoolScriptArg("false")
		require.NoError(t, err)
		args = append(args, arg)
	}
	{
		arg, err := vm_client.NewVectorScriptArg("0x0001")
		require.NoError(t, err)
		args = append(args, arg)
	}

	scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, args)
	require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "script execute error")

	checkNoEventErrors(input.ctx.EventManager().Events(), t)
}

// Test that all hardcoded VM Path are correct.
// If something goes wrong, check the DataSource logs for requested Path and fix.
func TestVMKeeper_Path(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	baseAmount := sdk.NewInt(1000)
	accCoins := sdk.NewCoins(
		sdk.NewCoin("xfi", baseAmount),
		sdk.NewCoin("eth", baseAmount),
		sdk.NewCoin("btc", baseAmount),
		sdk.NewCoin("usdt", baseAmount),
	)

	addr1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	_ = acc1.SetCoins(accCoins)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "MiddlewareBlock",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "MiddlewareTime",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check vmauth module path: XFI
	testID = "VMAuth XFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
    			use 0x1::Account;
				use 0x1::XFI;

				fun main(account: &signer) {
					let _ = Account::balance<XFI::T>(account);
				}
			}
		`
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "VMAuthXFI",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "VMAuthETH",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "VMAuthUSDT",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "VMAuthBTC",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}

	// Check currencies_register module path: XFI
	testID = "CurrencyInfo XFI"
	{
		t.Logf("%s: script compile", testID)
		scriptSrc := `
			script {
				use 0x1::Dfinance;
				use 0x1::XFI;

				fun main() {
					let _ = Dfinance::denom<XFI::T>();
				}
			}
		`
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "CurrencyInfoXFI",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "CurrencyInfoETH",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "CurrencyInfoUSDT",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		bytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "CurrencyInfoBTC",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, bytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, bytecode[0].ByteCode, nil)
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
		moduleBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: moduleSrc,
					Name: "AccountModule",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: module compile error", testID)
		require.Len(t, moduleBytecode, 1)

		t.Logf("%s: module deploy", testID)
		moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode[0].ByteCode)
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
		scriptBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "AccountModuleScript",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "%s: script compile error", testID)
		require.Len(t, scriptBytecode, 1)

		t.Logf("%s: script execute", testID)
		scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode[0].ByteCode, nil)
		require.NoErrorf(t, input.vk.ExecuteScript(input.ctx, scriptMsg), "%s: script execute error", testID)

		t.Logf("%s: checking script events", testID)
		checkNoEventErrors(input.ctx.EventManager().Events(), t)
	}
}

// VM Event.EventType string serialization test.
func TestVMKeeper_EventTypeSerialization(t *testing.T) {
	const moduleSrc = `
		module Foo {
		    struct FooEvent<T, VT> {
		        field_T:  T,
		        field_VT: VT
		    }
		
		    public fun NewFooEvent<T, VT>(account: &signer, arg_T: T, arg_VT: VT): FooEvent<T, VT> {
		        let fooEvent = FooEvent<T, VT> {
		            field_T:  arg_T,
		            field_VT: arg_VT
		        };
				
				0x1::Event::emit<bool>(account, true);
		
		        fooEvent
		    }
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::Foo;
			
			fun main(account: &signer) {
				// Event with single tag
				0x1::Event::emit<u8>(account, 128);
				
				// Event with single vector
				0x1::Event::emit<vector<u8>>(account, x"0102");
				
				// Two events:
				//   1. Module: single tag
				//   2. Script: generic struct with tag, vector
				let fooEvent = Foo::NewFooEvent<u64, vector<u8>>(account, 1000, x"0102");
				0x1::Event::emit<Foo::FooEvent<u64, vector<u8>>>(account, fooEvent);
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
	moduleBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: moduleSrc,
				Name: "moduleSrc",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")
	require.Len(t, moduleBytecode, 1)

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode[0].ByteCode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: scriptSrc,
				Name: "AccountModuleScript",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")
	require.Len(t, scriptBytecode, 1)

	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode[0].ByteCode, nil)
	resp, err := input.vk.ExecuteScriptNoProcessing(input.ctx, scriptMsg)
	require.NoErrorf(t, err, "script execute error")

	t.Logf("checking script events")
	checkNoEventErrors(input.ctx.EventManager().Events(), t)

	for idx, event := range resp.Events {
		t.Logf("VM Event #%d", idx)
		t.Log(types.StringifyVMEvent(event))

		t.Logf("Cosmos Event #%d", idx)
		cosmosEvent := types.NewMoveEvent(sdk.NewInfiniteGasMeter(), event)
		printEvent(cosmosEvent, t)
	}
}

// VM Event.EventType string serialization test with gas charged check.
func TestVMKeeper_EventTypeSerializationGas(t *testing.T) {
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
		
			public fun test(account: &signer) {
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
		
				0x1::Event::emit<D<C<B<A>>>>(account, d);
			}
		
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::GasEvent;

			fun main(account: &signer) {
				GasEvent::test(account);
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
	moduleBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: moduleSrc,
				Name: "moduleSrc",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")
	require.Len(t, moduleBytecode, 1)

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode[0].ByteCode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: scriptSrc,
				Name: "scriptSrc",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")
	require.Len(t, scriptBytecode, 1)

	gasMeter := sdk.NewGasMeter(100000)
	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode[0].ByteCode, nil)
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
func TestVMKeeper_EventTypeSerializationOutOfGas(t *testing.T) {
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
		
			public fun test(account: &signer) {
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
		
				0x1::Event::emit<M<V<Z<C<B<A>>>>>>(account, m);
			}
		
		}
	`
	const scriptSrcFmt = `
		script {
			use %s::OutOfGasEvent;

			fun main(account: &signer) {
				OutOfGasEvent::test(account);
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
	moduleBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: moduleSrc,
				Name: "moduleSrc",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "module compile error")
	require.Len(t, moduleBytecode, 1)

	moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode[0].ByteCode)
	require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")
	ctx, writeCtx := input.ctx.CacheContext()
	require.NoErrorf(t, input.vk.DeployContract(ctx, moduleMsg), "module deploy error")

	t.Logf("checking module events")
	checkNoEventErrors(ctx.EventManager().Events(), t)
	writeCtx()

	// Compile, execute script
	scriptSrc := fmt.Sprintf(scriptSrcFmt, addr1)
	scriptBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
		Units: []*compiler_grpc.CompilationUnit{
			{
				Text: scriptSrc,
				Name: "scriptSrc",
			},
		},
		Address: common_vm.Bech32ToLibra(addr1),
	})
	require.NoErrorf(t, err, "script compile error")
	require.Len(t, scriptBytecode, 1)

	scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode[0].ByteCode, nil)

	require.PanicsWithValue(t, sdk.ErrorOutOfGas{Descriptor: "event type processing"}, func() {
		_ = input.vk.ExecuteScript(input.ctx.WithGasMeter(sdk.NewGasMeter(100000)), scriptMsg)
	})
}

func TestResearch_LCS(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := newTestInput(false)

	// Create account
	var addr1 sdk.AccAddress
	var addr1Libra [20]byte
	{
		baseAmount := sdk.NewInt(1000)
		accCoins := sdk.NewCoins(
			sdk.NewCoin("xfi", baseAmount),
			sdk.NewCoin("eth", baseAmount),
			sdk.NewCoin("btc", baseAmount),
			sdk.NewCoin("usdt", baseAmount),
		)

		addr1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
		acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
		_ = acc1.SetCoins(accCoins)
		input.ak.SetAccount(input.ctx, acc1)

		copy(addr1Libra[:], common_vm.Bech32ToLibra(addr1)[:20])
		t.Logf("Address: 0x%s", hex.EncodeToString(addr1Libra[:]))
	}

	// Init genesis and start DS
	{
		gs := getGenesis(t)
		input.vk.InitGenesis(input.ctx, gs)
		input.vk.SetDSContext(input.ctx)
		input.vk.StartDSServer(input.ctx)
		time.Sleep(2 * time.Second)
	}

	// Launch DVM container
	stopContainer := startDVMContainer(t, input.dsPort)
	defer stopContainer()

	// Read and compile sources, deploy byteCodes
	{
		moduleSrc := readFile(t, "./move/lcs_module.move")
		moduleSrc = strings.ReplaceAll(moduleSrc, "0x123", addr1.String())
		moduleBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: moduleSrc,
					Name: "moduleSrc",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "module compile error")
		require.Len(t, moduleBytecode, 1)

		// publish
		{
			moduleMsg := types.NewMsgDeployModule(addr1, moduleBytecode[0].ByteCode)
			require.NoErrorf(t, moduleMsg.ValidateBasic(), "module deploy message validation failed")

			cacheCtx, writeCtx := input.ctx.CacheContext()
			require.NoErrorf(t, input.vk.DeployContract(cacheCtx, moduleMsg), "module deploy error")

			checkNoEventErrors(cacheCtx.EventManager().Events(), t)
			writeCtx()
		}

		scriptSrc := readFile(t, "./move/lcs_script.move")
		scriptSrc = strings.ReplaceAll(scriptSrc, "0x123", addr1.String())
		scriptBytecode, err := vm_client.Compile(*vmCompiler, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: scriptSrc,
					Name: "scriptSrc",
				},
			},
			Address: common_vm.Bech32ToLibra(addr1),
		})
		require.NoErrorf(t, err, "script compile error")
		require.Len(t, scriptBytecode, 1)

		// execute
		{
			scriptMsg := types.NewMsgExecuteScript(addr1, scriptBytecode[0].ByteCode, nil)
			require.NoErrorf(t, scriptMsg.ValidateBasic(), "script execute message validation failed")

			cacheCtx, writeCtx := input.ctx.CacheContext()
			require.NoErrorf(t, input.vk.ExecuteScript(cacheCtx, scriptMsg), "script execute error")

			checkNoEventErrors(cacheCtx.EventManager().Events(), t)
			writeCtx()
		}
	}

	// Build VM resource path
	resPath := glav.NewStructTag(addr1Libra, "Foo", "Bar", nil).AccessVector()
	t.Logf("Resource path: 0x%s", hex.EncodeToString(resPath))

	// Get resource raw data
	resAccessPath := &vm_grpc.VMAccessPath{Address: addr1, Path: resPath}
	resRawData := input.vk.GetValue(input.ctx, resAccessPath)
	require.NotNil(t, resRawData, "resource raw data not found")
	t.Logf("Raw resource data: %s", hex.EncodeToString(resRawData))

	// Viewer parsing
	viewerRequest := types.ViewerRequest{
		types.ViewerItem{Name: "u8Val", Type: "U8"},
		types.ViewerItem{Name: "u64Val", Type: "U64"},
		types.ViewerItem{Name: "u128Val", Type: "U128"},
		types.ViewerItem{Name: "boolVal", Type: "bool"},
		types.ViewerItem{Name: "addrVal", Type: "address"},
		types.ViewerItem{
			Name:      "vectU8Val",
			Type:      "vector",
			InnerItem: &types.ViewerRequest{types.ViewerItem{Type: "U8"}},
		},
		types.ViewerItem{
			Name:      "vectU64Val",
			Type:      "vector",
			InnerItem: &types.ViewerRequest{types.ViewerItem{Type: "U64"}},
		},
		types.ViewerItem{
			Name: "innerStruct",
			Type: "struct",
			InnerItem: &types.ViewerRequest{
				types.ViewerItem{Name: "a", Type: "U8"},
				types.ViewerItem{Name: "b", Type: "bool"},
			},
		},
		types.ViewerItem{
			Name: "vectComplex",
			Type: "vector",
			InnerItem: &types.ViewerRequest{
				types.ViewerItem{
					Type: "struct",
					InnerItem: &types.ViewerRequest{
						types.ViewerItem{Name: "a", Type: "U8"},
						types.ViewerItem{Name: "b", Type: "bool"},
					},
				},
			},
		},
	}
	resString, err := StringifyLCSData(viewerRequest, resRawData)
	require.NoError(t, err, "viewer parsing")
	require.NotEmpty(t, resString)
	t.Logf("Viewer result:\n%s", resString)
}
