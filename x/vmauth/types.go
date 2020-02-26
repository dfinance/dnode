package vmauth

import (
	"encoding/hex"
	"github.com/WingsDao/wings-blockchain/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/the729/lcs"
	"math/big"
	"strconv"
)

const (
	amountLength = 16
	denomLength  = 15
	resourceKey  = "016ee00e2d212d7676b19de9ce7a4b598a339ae2286ef6b378c0c348b3fd3221ed"
)

/*
// coin struct
struct Coin {
	denom: bytearray,
	value: u64,
}

resource T {
	// all balances.
	balances: Vector.T<Self.Coin>,
}
*/
type WBCoin struct {
	Denom []byte // coin denom
	Value uint64 // coin value
}

type AccountResource struct {
	Balances []WBCoin // coins
}

// Just convert little endian to big endian (because of Cosmos sdk usage of big endian).
func LeToBe(bytes []byte) {
	for i := 0; i < len(bytes)/2; i++ {
		bytes[len(bytes)-i-1], bytes[i] = bytes[i], bytes[len(bytes)-i-1]
	}
}

// Convert bytes slice to coin.
func bytesToCoin(bytes []byte) sdk.Coin {
	numBz := bytes[:amountLength]
	LeToBe(numBz)

	amount := &big.Int{}
	amount.SetBytes(numBz)

	denom := make([]byte, 0)
	for i := amountLength; i < amountLength+denomLength; i++ {
		if bytes[i] != 0 {
			denom = append(denom, bytes[i])
		} else {
			break
		}
	}

	return sdk.Coin{Amount: sdk.NewIntFromBigInt(amount), Denom: string(denom)}
}

// byte array to chunks.
func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

// convert byte array to coins.
func bytesToCoins(coins []WBCoin) sdk.Coins {
	if coins == nil {
		return nil
	} else {
		realCoins := make(sdk.Coins, 0)
		for _, coin := range coins {
			str := strconv.FormatUint(coin.Value, 10)
			amount, _ := sdk.NewIntFromString(str)
			realCoins = append(realCoins, sdk.NewCoin(string(coin.Denom), amount))
		}

		return realCoins
	}
}

// Convert coin to  bytes
func coinToBytes(coin sdk.Coin) []byte {
	// 128 bits (16 bytes) for balance
	// 15 bytes for denom
	// little endian for numbers
	val := helpers.BigToBytes(coin.Amount, amountLength)
	denom := make([]byte, denomLength)
	for i, c := range coin.Denom {
		denom[i] = byte(c)
	}
	return append(val, denom...)
}

// Convert coins to bytes.
func coinsToBytes(coins sdk.Coins) []byte {
	bytes := make([]byte, 0)
	for _, coin := range coins {
		if !coin.IsZero() {
			bytes = append(bytes, coinToBytes(coin)...)
		}
	}

	return bytes
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
				Value: coin.Amount.BigInt().Uint64(),
			})
		}

		return AccountResource{Balances: balances}
	}
}

// Convert acc to bytes
func AccToBytes(acc AccountResource) []byte {
	bytes, err := lcs.Marshal(acc)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BytesToAccRes(bz []byte) AccountResource {
	var accRes AccountResource
	err := lcs.Unmarshal(bz, &accRes)
	if err != nil {
		panic(err)
	}

	return accRes
}
