package vmauth

import (
	"encoding/binary"
	"encoding/hex"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/x/common_vm"

	"github.com/dfinance/lcs"
)

// Resource key for WBCoins resource from VM stdlib.
const (
	resourceKey   = "017f0e04d8f92bed6b87baff6145039aad7eb605e8b76e117f523fbd9079253d72"
	ehResourceKey = "01923cc4af19291acaf06a40e1d4f9f922af9ffc4ea263050c2f867145468613c4"
)

var (
	ErrInternal = sdkErrors.Register(auth.ModuleName, 100, "internal")
)

type EventHandleGenerator struct {
	Counter uint64
	Addr    []byte `lcs:"len=24"`
}

type DNCoin struct {
	Denom []byte   // coin denom
	Value *big.Int // coin value
}

// Event handle for account.
type EventHandle struct {
	Counter uint64
	Guid    []byte
}

// Balances of account in case of standard lib.
type AccountResource struct {
	Balances       []DNCoin     // coins.
	WithdrawEvents *EventHandle // receive events handler.
	DepositEvents  *EventHandle // sent events handler.
}

type EventGenerator struct {
	Counter uint64
	Address []byte
}

// Convert byte array to coins.
func balancesToCoins(coins []DNCoin) sdk.Coins {
	realCoins := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		realCoins[i] = sdk.NewCoin(string(coin.Denom), sdk.NewIntFromBigInt(coin.Value))
	}

	return realCoins
}

// Get resource path.
func GetResPath() []byte {
	data, err := hex.DecodeString(resourceKey)
	if err != nil {
		panic(err)
	}

	return data
}

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

// Convert sdk.Coins to resources.
func convertCoins(coins sdk.Coins) []DNCoin {
	balances := make([]DNCoin, len(coins))
	for i, coin := range coins {
		balances[i] = DNCoin{
			Denom: []byte(coin.Denom),
			Value: coin.Amount.BigInt(),
		}
	}

	return balances
}

// Merging exported.Account with VM account (resource), returns new AccountResource contains merged events handlers.
func MergeVMAccountEvents(acc exported.Account, source AccountResource) AccountResource {
	return AccountResource{
		Balances:       convertCoins(acc.GetCoins()),
		WithdrawEvents: source.WithdrawEvents,
		DepositEvents:  source.DepositEvents,
	}
}

// Creating new VM account and EventHandleGenerator.
func CreateVMAccount(acc exported.Account) (vmAcc AccountResource, eventHandleGen EventHandleGenerator) {
	vmAcc = AccountResource{
		Balances: convertCoins(acc.GetCoins()),
	}

	eventHandleGen = EventHandleGenerator{
		Counter: 0,
		Addr:    common_vm.Bech32ToLibra(acc.GetAddress()),
	}

	// just create new event handlers.
	vmAcc.WithdrawEvents = &EventHandle{
		Counter: 0,
		Guid:    getGUID(acc.GetAddress(), eventHandleGen.Counter),
	}

	eventHandleGen.Counter += 1

	vmAcc.DepositEvents = &EventHandle{
		Counter: 0,
		Guid:    getGUID(acc.GetAddress(), eventHandleGen.Counter),
	}

	eventHandleGen.Counter += 1
	return
}

// Event generator to bytes.
func EventHandlerGenToBytes(eh EventHandleGenerator) []byte {
	bytes, err := lcs.Marshal(eh)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Convert account resource to bytes.
func AccResToBytes(acc AccountResource) []byte {
	bytes, err := lcs.Marshal(acc)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("Write to path: %s\n", hex.EncodeToString(bytes))

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
