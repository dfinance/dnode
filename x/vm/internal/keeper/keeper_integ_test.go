// +build integ

package keeper

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"

	dnodeConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/vm/client/cli"
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

// Test transfer of dfi between two accounts in dfi.
func TestKeeper_DeployContractTransfer(t *testing.T) {
	config := sdk.GetConfig()
	dnodeConfig.InitBechPrefixes(config)

	input := setupTestInput(false)

	// launch docker
	client, c := launchDocker(t)
	t.Log("launched docker")
	defer stopDocker(t, client, c)

	// create accounts.
	addr1, err := sdk.AccAddressFromBech32("wallet14ng6lzsvyy26sxmujmjthvrjde8x6gkk2gzeft")
	if err != nil {
		t.Fatal(err)
	}
	acc1 := input.ak.NewAccountWithAddress(input.ctx, addr1)
	acc1.SetCoins(sdk.NewCoins(sdk.NewCoin("dfi", sdk.NewInt(100))))

	addr2, err := sdk.AccAddressFromBech32("wallet1cx0l8zhwjdxuvcvjsqew7770kef2uazslzs33y")
	if err != nil {
		t.Fatal(err)
	}
	acc2 := input.ak.NewAccountWithAddress(input.ctx, addr2)

	input.ak.SetAccount(input.ctx, acc1)
	input.ak.SetAccount(input.ctx, acc2)

	// write write set.
	gs := getGenesis(t)
	input.vk.InitGenesis(input.ctx, gs)
	input.vk.SetDSContext(input.ctx)
	input.vk.StartDSServer(input.ctx)
	time.Sleep(1 * time.Second)

	// wait for compiler
	if err := waitStarted(client, c.ID, 5*time.Second); err != nil {
		t.Fatalf("can't connect to docker dvm: %v", err)
	}

	// wait reachable compiler
	if err := waitReachable(*vmCompiler, 5*time.Second); err != nil {
		t.Fatalf("can't connect to docker compiler: %v", err)
	}

	bytecode, err := cli.Compile(*vmCompiler, &vm_grpc.MvIrSourceFile{
		Text:    sendScript,
		Address: []byte(addr1.String()),
		Type:    vm_grpc.ContractType_Script,
	})
	if err != nil {
		t.Fatalf("can't get code for send script: %v", err)
	}

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

	for _, event := range events {
		if event.Type == types.EventTypeError {
			t.Fatalf("should not contains error event: %s %s", event.Attributes[0].Key, event.Attributes[0].Value)
		}
	}

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
