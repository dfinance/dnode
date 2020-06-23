package vmauth

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"

	"github.com/dfinance/dnode/x/common_vm"
)

// BalanceResource is an account coins balance used by DVM.
type BalanceResource struct {
	Value *big.Int
}

// Balance is a BalanceResource object with meta data.
type Balance struct {
	accessPath *vm_grpc.VMAccessPath
	denom      string
	balance    BalanceResource
}

// Balance slice type.
type Balances []Balance

// loadAccessPaths builds Balances with access paths based on internal matching map.
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

// coinToBalance converts sdk.Coin for sdk.AccAddress to Balance.
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

// coinsToBalances converts account Coins to Balance resources.
// Returns two kind of Balances: to write and to delete.
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

// balanceToCoin converts Balance to sdk.Coin.
func balanceToCoin(balance Balance) sdk.Coin {
	return sdk.NewCoin(balance.denom, sdk.NewIntFromBigInt(balance.balance.Value))
}

// balancesToCoins converts Balances to sdk.Coins.
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

// BalanceToBytes marshals BalanceResource object.
func BalanceToBytes(balance BalanceResource) []byte {
	bytes, err := lcs.Marshal(balance)
	if err != nil {
		panic(err)
	}

	return bytes
}

// BytesToBalance unmarshals BalanceResource object.
func BytesToBalance(bz []byte) BalanceResource {
	var balance BalanceResource
	err := lcs.Unmarshal(bz, &balance)
	if err != nil {
		panic(err)
	}

	return balance
}
