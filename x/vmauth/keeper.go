// Implements account keeper with vm storage inside to allow work with account resources from VM.
package vmauth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/common_vm"
)

// Implements account keeper with vm storage support.
type VMAccountKeeper struct {
	*auth.AccountKeeper

	cdc      *codec.Codec
	vmKeeper common_vm.VMStorage
}

// Create new account vm keeper.
func NewVMAccountKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, vmKeeper common_vm.VMStorage, proto func() exported.Account) VMAccountKeeper {
	keeper := auth.NewAccountKeeper(cdc, key, paramstore, proto)

	return VMAccountKeeper{
		AccountKeeper: &keeper,
		vmKeeper:      vmKeeper,
		cdc:           cdc,
	}
}

// saveBalances updates / removes Balance resources from the storage.
func (keeper VMAccountKeeper) saveBalances(ctx sdk.Context, balances Balances, toDelete Balances) {
	for _, balance := range balances {
		keeper.vmKeeper.SetValue(ctx, balance.accessPath, BalanceToBytes(balance.balance))
	}

	for _, toDel := range toDelete {
		if keeper.vmKeeper.HasValue(ctx, toDel.accessPath) {
			keeper.vmKeeper.SetValue(ctx, toDel.accessPath, BalanceToBytes(toDel.balance))
		}
	}
}

// loadBalances gets Balance resources from the storage.
func (keeper VMAccountKeeper) loadBalances(ctx sdk.Context, addr sdk.AccAddress) Balances {
	balances := loadAccessPaths(addr)
	realBalances := make([]Balance, 0)

	for _, balance := range balances {
		bz := keeper.vmKeeper.GetValue(ctx, balance.accessPath)
		if bz != nil {
			balanceRes := BytesToBalance(bz)
			balance.balance = balanceRes
			realBalances = append(realBalances, balance)
		}
	}

	return realBalances
}

// SetAccount stores account resources to the storage and updates std keeper.
func (keeper VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	// Update balances extracted from Coins.
	balances, toDelete := coinsToBalances(acc)
	keeper.saveBalances(ctx, balances, toDelete)

	// Add account to std keeper.
	keeper.AccountKeeper.SetAccount(ctx, acc)
}

// GetAccount reads account from the std keeper and updates account resources.
func (keeper VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	// Get account from std keeper.
	account := keeper.AccountKeeper.GetAccount(ctx, addr)

	// Update balances.
	balances := keeper.loadBalances(ctx, addr)
	if len(balances) > 0 {
		realCoins := balancesToCoins(balances)

		if account != nil {
			if !realCoins.IsEqual(account.GetCoins()) {
				if err := account.SetCoins(realCoins); err != nil {
					panic(err) // must never happen
				}

				keeper.SetAccount(ctx, account)
			}
		} else {
			account = keeper.NewAccountWithAddress(ctx, addr)
			if err := account.SetCoins(realCoins); err != nil {
				panic(err) // should never happen
			}

			keeper.SetAccount(ctx, account)
		}
	}

	return account
}

// GetAllAccounts returns all accounts in the std keeper.
// As it's not colled anywhere (as it seems), we can ignore vm storage for now.
// todo: process all vm storage accounts and compare with standard accounts.
func (keeper VMAccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	accounts := keeper.AccountKeeper.GetAllAccounts(ctx)

	return accounts
}

// RemoveAccount removes account resources from the storage and removes account from the std keeper.
// NOTE: this will cause supply invariant violation if called.
func (keeper VMAccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	// Remove account from the std keeper.
	keeper.AccountKeeper.RemoveAccount(ctx, acc)

	// Remove all Balance resources.
	balances := loadAccessPaths(acc.GetAddress())
	for _, b := range balances {
		keeper.vmKeeper.DelValue(ctx, b.accessPath)
	}
}

// GetSignerAcc returns an account for a given address that is expected to sign a transaction.
func GetSignerAcc(ctx sdk.Context, ak VMAccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q does not exist", addr)
}
