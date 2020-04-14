package vmauth

import (
	"encoding/binary"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/helpers"
)

// Resource key for WBCoins resource from VM stdlib.
const (
	resourceKey        = "01fc0e661c5c73d4acaf1c8d0494acec68953a8279674d9e850fc11f36b31302c2"
	libraAddressLength = 32
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
	Count uint64
	Key   []byte
}

// Balances of account in case of standard lib.
type AccountResource struct {
	Balances       []DNCoin     // coins.
	WithdrawEvents *EventHandle // receive events handler.
	DepositEvents  *EventHandle // sent events handler.
	EventGenerator uint64       // event generator.
}

// Convert byte array to coins.
func balancesToCoins(coins []DNCoin) sdk.Coins {
	realCoins := make(sdk.Coins, len(coins))
	for i, coin := range coins {
		realCoins[i] = sdk.NewCoin(string(coin.Denom), coin.Value)
	}

	return realCoins
}

// Bytes to libra compability.
func AddrToPathAddr(addr []byte) []byte {
	config := sdk.GetConfig()
	prefix := config.GetBech32AccountAddrPrefix()
	zeros := make([]byte, libraAddressLength-len(prefix)-len(addr))

	bytes := make([]byte, 0)
	bytes = append(bytes, []byte(prefix)...)
	bytes = append(bytes, zeros...)
	bytes = append(bytes, addr...)

	return bytes
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
	bzCounter := make([]byte, 8)

	addr := AddrToPathAddr(address)

	binary.LittleEndian.PutUint64(bzCounter, counter)

	return append(bzCounter, addr...)
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
		accRes.EventGenerator = source.EventGenerator
	} else {
		// just create new event handlers.
		accRes.EventGenerator = 0

		accRes.WithdrawEvents = &EventHandle{
			Count: 0,
			Key:   getGUID(acc.GetAddress(), accRes.EventGenerator),
		}
		accRes.EventGenerator += 1

		//  increase event generator for another id.
		accRes.DepositEvents = &EventHandle{
			Count: 0,
			Key:   getGUID(acc.GetAddress(), accRes.EventGenerator),
		}
		accRes.EventGenerator += 1
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
