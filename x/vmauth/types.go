package vmauth

import (
	"encoding/binary"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/common_vm"
)

// Resource key for WBCoins resource from VM stdlib.
const (
	resourceKey = "017f0e04d8f92bed6b87baff6145039aad7eb605e8b76e117f523fbd9079253d72"
)

var (
	ErrInternal = sdkErrors.Register(auth.ModuleName, 100, "internal")
)

type DNCoin struct {
	Denom []byte  // coin denom
	Value sdk.Int // coin value
}

// Event handle for account.
type EventHandle struct {
	Counter uint64
	Guid    []byte
}

// Balances of account in case of standard lib.
type AccountResource struct {
	A              uint64
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
		realCoins[i] = sdk.NewCoin(string(coin.Denom), coin.Value)
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

// Get GUID for events.
func getGUID(address sdk.AccAddress, counter uint64) []byte {
	/*
		let sender_bytes = LCS::to_bytes(&counter.addr);
		let count_bytes = LCS::to_bytes(&counter.counter);
		counter.counter = counter.counter + 1;

		// EventHandleGenerator goes first just in case we want to extend address in the future.
		Vector::append(&mut count_bytes, sender_bytes);

		count_bytes
	*/
	senderBytes, err := helpers.Marshal(common_vm.Bech32ToLibra(address))
	if err != nil {
		panic(err) // should not happen
	}

	countBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(countBytes, counter)
	return append(countBytes, senderBytes...)
}

// Convert acc to account resource.
func AccResFromAccount(acc exported.Account, source *AccountResource) AccountResource {
	accCoins := acc.GetCoins()
	balances := make([]DNCoin, len(accCoins))
	for i, coin := range accCoins {
		balances[i] = DNCoin{
			Denom: []byte(coin.Denom),
			Value: coin.Amount,
		}
	}

	accRes := AccountResource{
		Balances: balances,
	}

	if source != nil {
		accRes.WithdrawEvents = source.WithdrawEvents
		accRes.DepositEvents = source.DepositEvents

		// also recopy event generator
		//accRes.EventGenerator = source.EventGenerator
	} else {
		// just create new event handlers.
		var generator uint64 = 0

		accRes.WithdrawEvents = &EventHandle{
			Counter: 0,
			Guid:    getGUID(acc.GetAddress(), generator),
		}
		generator += 1

		//  increase event generator for another id.
		accRes.DepositEvents = &EventHandle{
			Counter: 0,
			Guid:    getGUID(acc.GetAddress(), generator),
		}

		generator += 1
	}

	return accRes
}

// Convert account resource to bytes.
func AccResToBytes(acc AccountResource) []byte {
	bytes, err := helpers.Marshal(acc)
	if err != nil {
		panic(err)
	}

	return bytes
}

// Unmarshall bytes to account.
func BytesToAccRes(bz []byte) AccountResource {
	var accRes AccountResource
	err := helpers.Unmarshal(bz, &accRes)
	if err != nil {
		panic(err)
	}

	return accRes
}
