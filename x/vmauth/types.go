package vmauth

import (
	"encoding/hex"
	"github.com/WingsDao/wings-blockchain/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
)

const (
	amountLength = 16
	resourceKey  = "016ee00e2d212d7676b19de9ce7a4b598a339ae2286ef6b378c0c348b3fd3221ed"
)

type WBCoin struct {
	Denom []byte  // coin denom
	Value sdk.Int // coin value
}

type AccountResource struct {
	Balances []WBCoin // coins
}

// convert byte array to coins.
func bytesToCoins(coins []WBCoin) sdk.Coins {
	if coins == nil {
		return nil
	} else {
		realCoins := make(sdk.Coins, 0)
		for _, coin := range coins {
			realCoins = append(realCoins, sdk.NewCoin(string(coin.Denom), coin.Value))
		}

		return realCoins
	}
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
	if acc.GetCoins() == nil {
		return AccountResource{}
	} else {
		accCoins := acc.GetCoins()
		balances := make([]WBCoin, 0)
		for _, coin := range accCoins {
			balances = append(balances, WBCoin{
				Denom: []byte(coin.Denom),
				Value: coin.Amount,
			})
		}

		return AccountResource{Balances: balances}
	}
}

// Convert acc to bytes
func AccToBytes(acc AccountResource) []byte {
	bytes, err := helpers.Marshal(acc)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BytesToAccRes(bz []byte) AccountResource {
	var accRes AccountResource
	err := helpers.Unmarshal(bz, &accRes)
	if err != nil {
		panic(err)
	}

	return accRes
}
