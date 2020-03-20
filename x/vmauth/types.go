package vmauth

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/helpers"
)

// Resource key for WBCoins resource from VM stdlib.
const (
	resourceKey        = "01fc0e661c5c73d4acaf1c8d0494acec68953a8279674d9e850fc11f36b31302c2"
	libraAddressLength = 32
)

type DNCoin struct {
	Denom []byte  // coin denom
	Value sdk.Int // coin value
}

// Balances of account in case of standard lib.
type AccountResource struct {
	Balances []DNCoin // coins
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

// Convert acc to account resource.
func AccResFromAccount(acc exported.Account) AccountResource {
	accCoins := acc.GetCoins()
	balances := make([]DNCoin, len(accCoins))
	for i, coin := range accCoins {
		balances[i] = DNCoin{
			Denom: []byte(coin.Denom),
			Value: coin.Amount,
		}
	}

	return AccountResource{Balances: balances}
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
