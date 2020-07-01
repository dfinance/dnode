// Implements account keeper with vm storage inside to allow work with account resources from VM.
package vmauth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"

	ccTypes "github.com/dfinance/dnode/x/currencies"
)

// Implements account keeper with vm storage support.
type VMAccountKeeper struct {
	*auth.AccountKeeper

	cdc      *codec.Codec
	ccKeeper ccTypes.Keeper
}

// Create new account vm keeper.
func NewVMAccountKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, proto func() exported.Account) *VMAccountKeeper {
	authKeeper := auth.NewAccountKeeper(cdc, key, paramstore, proto)

	return &VMAccountKeeper{
		cdc:           cdc,
		AccountKeeper: &authKeeper,
	}
}

func (k *VMAccountKeeper) SetCurrenciesKeeper(ccKeeper ccTypes.Keeper) {
	k.ccKeeper = ccKeeper
}

// SetAccount stores account resources to VM storage and updates std keeper.
func (k *VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	// update balances extracted from account coins
	if err := k.ccKeeper.SetAccountBalanceResources(ctx, acc); err != nil {
		panic(err)
	}

	// add account to std keeper
	k.AccountKeeper.SetAccount(ctx, acc)
}

// GetAccount reads account from the std keeper and updates account resources.
func (k *VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	// get account from std keeper
	acc := k.AccountKeeper.GetAccount(ctx, addr)

	// get balance resources
	balances, err := k.ccKeeper.GetAccountBalanceResources(ctx, addr)
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
// todo: process all vm storage accounts and compare with standard accounts.
func (k *VMAccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	return k.AccountKeeper.GetAllAccounts(ctx)
}

// RemoveAccount removes account resources from the storage and removes account from the std keeper.
// NOTE: this will cause supply invariant violation if called.
func (k *VMAccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	// remove account from the std keeper
	k.AccountKeeper.RemoveAccount(ctx, acc)

	// remove all account resources
	k.ccKeeper.RemoveAccountBalanceResources(ctx, acc.GetAddress())
}

// GetSignerAcc returns an account for a given address that is expected to sign a transaction.
func GetSignerAcc(ctx sdk.Context, ak *VMAccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q does not exist", addr)
}
