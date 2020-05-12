package keeper

import (
	"encoding/binary"
	"encoding/hex"
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
use 0x0::Account;
use 0x0::Coins;
use 0x0::DFI;

fun main(recipient: address, amount: u128) {
    Account::pay_from_sender<DFI::T>(recipient, amount);
    Account::pay_from_sender<Coins::ETH>(recipient, amount);
    Account::pay_from_sender<Coins::BTC>(recipient, amount);
    Account::pay_from_sender<Coins::USDT>(recipient, amount);
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
use 0x0::Event;
use {{sender}}::Math;

fun main(a: u64, b: u64) {
	let c = Math::add(a, b);

	let event_handle = Event::new_event_handle<u64>();
	Event::emit_event(&mut event_handle, c);
	Event::destroy_handle(event_handle);
}
`

const oraclePriceScript = `
use 0x0::Event;
use 0x0::Oracle;

fun main(ticket: u64) {
    let price = Oracle::get_price(ticket);

    let event_handle = Event::new_event_handle<u64>();
	Event::emit_event(&mut event_handle, price);
	Event::destroy_handle(event_handle);
}
`

func printEvent(event sdk.Event, t *testing.T) {
	t.Logf("Event: %s\n", event.Type)
	for _, attr := range event.Attributes {
		t.Logf("%s = %s\n", attr.Key, attr.Value)
	}
}

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
	toSend := sdk.NewInt(100)
	putCoins := sdk.NewCoins(
		sdk.NewCoin("dfi", baseAmount),
		sdk.NewCoin("eth", baseAmount),
		sdk.NewCoin("btc", baseAmount),
		sdk.NewCoin("usdt", baseAmount),
	)

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

	bytecode, err := compilerClient.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    sendScript,
		Address: common_vm.Bech32ToLibra(addr1),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for send script: %v", err)

	// execute contract.
	args := make([]types.ScriptArg, 2)
	args[0] = types.ScriptArg{
		Value: addr2.String(),
		Type:  vm_grpc.VMTypeTag_Address,
	}
	args[1] = types.ScriptArg{
		Value: toSend.String(),
		Type:  vm_grpc.VMTypeTag_U128,
	}

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
		require.Equal(t, baseAmount.Sub(toSend).String(), got.Amount.String())
	}

	recipient := input.ak.GetAccount(input.ctx, addr2)
	recpCoins := recipient.GetCoins()

	for _, got := range recpCoins {
		require.Equal(t, toSend.String(), got.Amount.String())
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

	require.Equal(t, events[1].Type, types.EventTypeMvirEvent, "script after execution doesn't contain event with amount")

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
