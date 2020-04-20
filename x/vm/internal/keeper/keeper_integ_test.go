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
	"github.com/dfinance/dnode/x/pricefeed"
	compilerClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const sendScript = `
import 0x0.Account;
import 0x0.Coins;

main(recipient: address, amount: u128, denom: bytearray) {
    let coin: Coins.Coin;
    coin = Account.withdraw_from_sender(move(amount), move(denom));

    Account.deposit(move(recipient), move(coin));
    return;
}
`

const mathModule = `
module Math {
    public add(a: u64, b: u64): u64 {
        return move(a) + move(b);
    }
}
`

const mathScript = `
import 0x0.Account;
import {{sender}}.Math;

main(a: u64, b: u64) {
	let c: u64;
	let handle: Account.EventHandle<u64>;

	c = Math.add(move(a), move(b));

    handle = Account.new_event_handle<u64>();
    Account.emit_event<u64>(&mut handle, move(c));
    Account.destroy_handle<u64>(move(handle));

	return;
}
`

const oraclePriceScript = `
import 0x0.Account;
import 0x0.Oracle;

main(ticket: u64) {
    let price: u64;
    let handle: Account.EventHandle<u64>;

    price = Oracle.get_price(move(ticket));

    handle = Account.new_event_handle<u64>();
    Account.emit_event<u64>(&mut handle, move(price));
    Account.destroy_handle<u64>(move(handle));
    return;
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
	acc1.SetCoins(sdk.NewCoins(sdk.NewCoin("dfi", sdk.NewInt(100))))

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
		Address: []byte(addr1.String()),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for send script: %v", err)

	// execute contract.
	args := make([]types.ScriptArg, 3)
	args[0] = types.ScriptArg{
		Value: addr2.String(),
		Type:  vm_grpc.VMTypeTag_Address,
	}
	args[1] = types.ScriptArg{
		Value: "100",
		Type:  vm_grpc.VMTypeTag_U128,
	}
	args[2] = types.ScriptArg{
		Value: fmt.Sprintf("b\"%s\"", hex.EncodeToString([]byte("dfi"))),
		Type:  vm_grpc.VMTypeTag_ByteArray,
	}

	msgScript := types.NewMsgExecuteScript(addr1, bytecode, args)
	err = input.vk.ExecuteScript(input.ctx, msgScript)
	require.NoError(t, err)

	events := input.ctx.EventManager().Events()
	require.Contains(t, events, types.NewEventKeep())

	checkNoErrors(events, t)

	// check balance changes
	sender := input.ak.GetAccount(input.ctx, addr1)
	coins := sender.GetCoins()

	for _, coin := range coins {
		if coin.Denom == "dfi" {
			require.Zero(t, coin.Amount.Int64())
		}
	}

	recipient := input.ak.GetAccount(input.ctx, addr2)
	require.Contains(t, recipient.GetCoins(), sdk.NewCoin("dfi", sdk.NewInt(100)))
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
		Address: []byte(addr1.String()),
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
		Address: []byte(addr1.String()),
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
	okInitParams := pricefeed.Params{
		Assets: pricefeed.Assets{
			pricefeed.Asset{
				AssetCode: assetCode,
				PriceFeeds: pricefeed.PriceFeeds{
					pricefeed.PriceFeed{
						Address: addr1,
					},
				},
				Active: true,
			},
		},
		Nominees: []string{addr1.String()},
		PostPrice: pricefeed.PostPriceParams{
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
		Address: []byte(addr1.String()),
		Type:    vm_grpc.ContractType_Script,
	})
	require.NoErrorf(t, err, "can't get code for price feed script: %v", err)

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
