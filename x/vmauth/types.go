package vmauth

import (
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/helpers"
)

const (
	resourceKey = "016ee00e2d212d7676b19de9ce7a4b598a339ae2286ef6b378c0c348b3fd3221ed"
)

type DNCoin struct {
	Denom []byte  // coin denom
	Value sdk.Int // coin value
}

type AccountResource struct {
	Balances []DNCoin // coins
}

// convert byte array to coins.
func bytesToCoins(coins []DNCoin) sdk.Coins {
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
	zeros := make([]byte, 5)

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
func AccResourceFromAccount(acc exported.Account) AccountResource {
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
func AccToBytes(acc AccountResource) []byte {
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
