package vmauth

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"

	"github.com/dfinance/lcs"
)

// Resource key for WBCoins resource from VM stdlib.
const (
	resourceKey   = "018fa369c050b9c85cae97841a141f9c622bc26673a88d083222ca87c36c6f5178"
	ehResourceKey = "01bb4c62af96260557ef1cb64dc0b265b5ebd765841d9d3f38b99a80feff5ec709"
)

var (
	// Errors.
	ErrInternal = sdkErrors.Register(auth.ModuleName, 100, "internal")
	denomPaths  map[string][]byte // path to denom.
)

// initialize denoms.
func init() {
	denomPaths = make(map[string][]byte)

	err := AddDenomPath("dfi", "01cd49bb0514afa42feffa4262acde3eb4415357ab68960754d5d8c1adf9e2acaf")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("eth", "01342063272df08689be301854f3bd18397d4962852d42e60e344e333ac71b2d83")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("usdt", "01afe7e50352ce4c687e0fb907fda30bace5304f4cea9e2a0eb674b14a29e40859")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("btc", "0122cc7d02ca9ad2fa853e31f68b44cacf30450a586613d4435e34b5054083872a")
	if err != nil {
		panic(err)
	}
}

func AddDenomPath(denom string, path string) error {
	var err error
	denomPaths[denom], err = hex.DecodeString(path)
	return err
}

func RemoveDenomPath(denom string) {
	delete(denomPaths, denom)
}

// Event generator for address.
type EventHandleGenerator struct {
	Counter uint64
	Addr    []byte `lcs:"len=20"`
}

// Balance.
type BalanceResource struct {
	Value *big.Int
}

// All balance resources.
type Balance struct {
	accessPath *vm_grpc.VMAccessPath
	denom      string
	balance    BalanceResource
}

// Balances type (contains several balances).
type Balances []Balance

// Event handle for account.
type EventHandle struct {
	Counter uint64
	Guid    []byte
}

// Balances of account in case of standard lib.
type AccountResource struct {
	SentEvents     *EventHandle // sent events handler.
	ReceivedEvents *EventHandle // received events handler.
}

// Load access paths for balances.
func loadAccessPaths(addr sdk.AccAddress) Balances {
	balances := make(Balances, len(denomPaths))

	i := 0
	for key, value := range denomPaths {
		accessPath := &vm_grpc.VMAccessPath{
			Address: common_vm.Bech32ToLibra(addr),
			Path:    value,
		}

		balances[i] = Balance{
			accessPath: accessPath,
			denom:      key,
		}

		i++
	}

	return balances
}

// Convert sdk.Coin to balance.
func coinToBalance(addr sdk.AccAddress, coin sdk.Coin) (Balance, error) {
	path, ok := denomPaths[coin.Denom]
	if !ok {
		return Balance{}, fmt.Errorf("cant find path for denom %s", coin.Denom)
	}

	return Balance{
		accessPath: &vm_grpc.VMAccessPath{
			Address: common_vm.Bech32ToLibra(addr),
			Path:    path,
		},
		denom: coin.Denom,
		balance: BalanceResource{
			Value: coin.Amount.BigInt(),
		},
	}, nil
}

// Convert coins to balances resources.
// Returns two kind of balances - to write and to delete.
func coinsToBalances(acc exported.Account) (Balances, Balances) {
	coins := acc.GetCoins()
	balances := make(Balances, len(coins))
	found := make(map[string]bool)

	for i, coin := range coins {
		var err error
		balances[i], err = coinToBalance(acc.GetAddress(), coin)
		if err != nil {
			panic(err)
		}
		found[coin.Denom] = true
	}

	toDelete := make(Balances, 0)
	for k := range denomPaths {
		if !found[k] {
			balance, err := coinToBalance(acc.GetAddress(), sdk.NewCoin(k, sdk.ZeroInt()))
			if err != nil {
				panic(err)
			}
			toDelete = append(toDelete, balance)
		}
	}

	return balances, toDelete
}

// Convert balance to sdk.Coin.
func balanceToCoin(balance Balance) sdk.Coin {
	return sdk.NewCoin(balance.denom, sdk.NewIntFromBigInt(balance.balance.Value))
}

// Convert balances to sdk.Coins.
func balancesToCoins(balances Balances) sdk.Coins {
	coins := make(sdk.Coins, 0)

	// if zero ignore return
	for _, balance := range balances {
		if balance.balance.Value.Cmp(sdk.ZeroInt().BigInt()) != 0 {
			coins = append(coins, balanceToCoin(balance))
		}
	}

	return coins
}

// Get resource path.
func GetResPath() []byte {
	data, err := hex.DecodeString(resourceKey)
	if err != nil {
		panic(err)
	}

	return data
}

// Get event handler generator resource path.
func GetEHPath() []byte {
	data, err := hex.DecodeString(ehResourceKey)
	if err != nil {
		panic(err)
	}

	return data
}

// Get GUID for events.
func getGUID(address sdk.AccAddress, counter uint64) []byte {
	countBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(countBytes, counter)
	return append(countBytes, common_vm.Bech32ToLibra(address)...)
}

// Creating new VM account and EventHandleGenerator.
func CreateVMAccount(acc exported.Account) (vmAcc AccountResource, eventHandleGen EventHandleGenerator) {
	vmAcc = AccountResource{}

	eventHandleGen = EventHandleGenerator{
		Counter: 0,
		Addr:    common_vm.Bech32ToLibra(acc.GetAddress()),
	}

	// just create new event handlers.
	vmAcc.SentEvents = &EventHandle{
		Counter: 0,
		Guid:    getGUID(acc.GetAddress(), eventHandleGen.Counter),
	}

	eventHandleGen.Counter += 1

	vmAcc.ReceivedEvents = &EventHandle{
		Counter: 0,
		Guid:    getGUID(acc.GetAddress(), eventHandleGen.Counter),
	}

	eventHandleGen.Counter += 1
	return
}

// Convert bytes to event handler generator.
func BytesToEventHandlerGen(bz []byte) EventHandleGenerator {
	var eventHandleGen EventHandleGenerator

	if err := lcs.Unmarshal(bz, &eventHandleGen); err != nil {
		panic(err)
	}
	return eventHandleGen
}

// Event handler generator to bytes.
func EventHandlerGenToBytes(eh EventHandleGenerator) []byte {
	bytes, err := lcs.Marshal(eh)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Convert balance resource to bytes.
func BalanceToBytes(balance BalanceResource) []byte {
	bytes, err := lcs.Marshal(balance)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Convert bytes to balances.
func BytesToBalance(bz []byte) BalanceResource {
	var balance BalanceResource
	err := lcs.Unmarshal(bz, &balance)
	if err != nil {
		panic(err)
	}

	return balance
}

// Convert account resource to bytes.
func AccResToBytes(acc AccountResource) []byte {
	bytes, err := lcs.Marshal(acc)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Unmarshall bytes to account.
func BytesToAccRes(bz []byte) AccountResource {
	var accRes AccountResource
	err := lcs.Unmarshal(bz, &accRes)
	if err != nil {
		panic(err)
	}

	return accRes
}
