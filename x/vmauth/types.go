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
	resourceKey   = "018a648879ad9e0d7db32230eaed39c4130bb965343b61fbd7e24b688413d51265"
	ehResourceKey = "011070ede6471d42ed53c740791479ad52b647b85c004b7e2559dd8bb8a024483d"
)

var (
	// Errors.
	ErrInternal = sdkErrors.Register(auth.ModuleName, 100, "internal")
	denomPaths  map[string][]byte // path to denom.
)

// initialize denoms.
func init() {
	denomPaths = make(map[string][]byte)

	err := AddDenomPath("dfi", "01ce4bdcb08e5d54437f1cb7e7f4a8cef079325a2f3daa8e7137e63b7cd22888a4")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("eth", "01e668bf6a3511e57c0f01096dccdec0ce4a755ba69612b2520ced925ad587bd62")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("usdt", "014c74bbc4ef1d30a77c63885fed3ca60cf15271a78926db9ff4014d6c084ee715")
	if err != nil {
		panic(err)
	}

	err = AddDenomPath("btc", "01b7b9499bf8ae1beaec05273475576d35e9a814a17c56a5d30342ee1912557f8f")
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
