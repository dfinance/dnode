// VM auth module keeper implements account keeper with additional VM resource handling.
package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/ccstorage"
)

// Module keeper object.
type VMAccountKeeper struct {
	auth.AccountKeeper

	cdc       *codec.Codec
	ccsKeeper ccstorage.Keeper
}

// SetAccount stores account resources to VM storage and updates std keeper.
func (k VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	if stdAcc := k.AccountKeeper.GetAccount(ctx, acc.GetAddress()); stdAcc == nil {
		for _, coin := range acc.GetCoins() {
			if err := k.ccsKeeper.IncreaseCurrencySupply(ctx, coin); err != nil {
				panic(fmt.Errorf("increasing currency %q supply for new account %q: %v", coin.Denom, acc.GetAddress(), err))
			}
		}
	}

	// update balances extracted from account coins
	if err := k.ccsKeeper.SetAccountBalanceResources(ctx, acc); err != nil {
		panic(err)
	}

	// add account to std keeper
	k.AccountKeeper.SetAccount(ctx, acc)
}

// GetAccount reads account from the std keeper and updates account resources.
func (k VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	// get account from std keeper
	acc := k.AccountKeeper.GetAccount(ctx, addr)

	// get balance resources
	balances, err := k.ccsKeeper.GetAccountBalanceResources(ctx, addr)
	if err != nil {
		panic(err)
	}

	// update std keeper's account and balance resources
	if len(balances) > 0 {
		realCoins := balances.Coins()
		if acc != nil {
			// account found (not newly created)
			if !realCoins.IsEqual(acc.GetCoins()) {
				// balances have changed, update account and resources
				if err := acc.SetCoins(realCoins); err != nil {
					panic(err)
				}
				k.SetAccount(ctx, acc)
			}
		} else {
			// new account, set account and resources
			acc = k.NewAccountWithAddress(ctx, addr)
			if err := acc.SetCoins(realCoins); err != nil {
				panic(err)
			}
			k.SetAccount(ctx, acc)
		}
	}

	return acc
}

// GetAllAccounts returns all accounts in the std keeper.
// As it's not called anywhere (as it seems), we can ignore VM storage for now.
// TODO: process all VM storage accounts and compare with standard accounts.
func (k VMAccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	return k.AccountKeeper.GetAllAccounts(ctx)
}

// RemoveAccount removes account resources from the storage and removes account from the std keeper.
// NOTE: this will cause supply invariant violation if called.
func (k VMAccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	// remove account from the std keeper
	k.AccountKeeper.RemoveAccount(ctx, acc)

	// remove all account resources
	k.ccsKeeper.RemoveAccountBalanceResources(ctx, acc.GetAddress())
}

// Create new keeper.
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, ccsKeeper ccstorage.Keeper, proto func() exported.Account) VMAccountKeeper {
	authKeeper := auth.NewAccountKeeper(cdc, key, paramstore, proto)

	return VMAccountKeeper{
		cdc:           cdc,
		AccountKeeper: authKeeper,
		ccsKeeper:     ccsKeeper,
	}
}

// GetSignerAcc returns an account for a given address that is expected to sign a transaction.
func GetSignerAcc(ctx sdk.Context, ak VMAccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q: not found", addr)
}
