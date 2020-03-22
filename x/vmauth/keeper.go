// Implements account keeper with vm storage inside to allow work with accounts from VM.
package vmauth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/vm"
)

// Implements account keeper with vm storage support.
type VMAccountKeeper struct {
	*auth.AccountKeeper

	cdc      *codec.Codec
	vmKeeper vm.VMStorage
}

// Create new account vm keeper.
func NewVMAccountKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, vmKeeper vm.VMStorage, proto func() exported.Account) VMAccountKeeper {
	keeper := auth.NewAccountKeeper(cdc, key, paramstore, proto)

	return VMAccountKeeper{
		AccountKeeper: &keeper,
		vmKeeper:      vmKeeper,
		cdc:           cdc,
	}
}

// Set account in storage.
func (keeper VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	keeper.AccountKeeper.SetAccount(ctx, acc)
	// now store account to vm storage
	accRes := AccResFromAccount(acc)
	keeper.vmKeeper.SetValue(ctx, &vm.VMAccessPath{
		Address: AddrToPathAddr(acc.GetAddress()),
		Path:    GetResPath(),
	}, AccResToBytes(accRes))
}

// Get account from storage.
func (keeper VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	account := keeper.AccountKeeper.GetAccount(ctx, addr)

	// check if account maybe exists in vm storage.
	bz := keeper.vmKeeper.GetValue(ctx, &vm.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	})

	// if account exists in vm.
	if bz != nil {
		accRes := BytesToAccRes(bz)
		realCoins := balancesToCoins(accRes.Balances)

		// load vm account from storage.
		// check if account exists in vm but not exists in our storage - if so, save account and return.
		// check if account has differences - balances, something else, and if so - save account and return.
		if account != nil {
			if !realCoins.IsEqual(account.GetCoins()) { // also check coins
				if err := account.SetCoins(realCoins); err != nil {
					helpers.CrashWithError(err) // should never happen
				}

				keeper.SetAccount(ctx, account)
			}
		} else {
			// if account is not exists - so create it.
			account = keeper.NewAccountWithAddress(ctx, addr)
			if err := account.SetCoins(realCoins); err != nil {
				helpers.CrashWithError(err) // should never happen
			}

			keeper.SetAccount(ctx, account)
		}
	}

	return account
}

// GetAllAccounts returns all accounts in the accountKeeper.
// as it's not calling anywhere, as it seems, we can ignore vm storage for now.
// todo: process all vm storage accounts and compare with standard accounts.
func (keeper VMAccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	accounts := keeper.AccountKeeper.GetAllAccounts(ctx)

	return accounts
}

// Removes an account from storage.
// NOTE: this will cause supply invariant violation if called
func (keeper VMAccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	keeper.AccountKeeper.RemoveAccount(ctx, acc)

	// should be remove account from VM storage too
	keeper.vmKeeper.DelValue(ctx, &vm.VMAccessPath{
		Address: AddrToPathAddr(acc.GetAddress()),
		Path:    GetResPath(),
	})
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak VMAccountKeeper, addr sdk.AccAddress) (exported.Account, sdk.Result) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, sdk.Result{}
	}

	return nil, sdk.ErrUnknownAddress(fmt.Sprintf("account %q does not exist", addr)).Result()
}
