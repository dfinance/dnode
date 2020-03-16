package vmauth

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"

	"github.com/dfinance/dnode/x/vm"
)

type VMAccountKeeper struct {
	*auth.AccountKeeper

	cdc      *codec.Codec
	vmKeeper vm.VMStorage
}

// Create new VM keeper.
func NewVMAccountKeeper(cdc *codec.Codec, key sdk.StoreKey, paramstore params.Subspace, vmKeeper vm.VMStorage, proto func() exported.Account) VMAccountKeeper {
	keeper := auth.NewAccountKeeper(cdc, key, paramstore, proto)

	return VMAccountKeeper{
		AccountKeeper: &keeper,
		vmKeeper:      vmKeeper,
		cdc:           cdc,
	}
}

func (keeper VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	keeper.AccountKeeper.SetAccount(ctx, acc)
	// now store account to vm storage
	accRes := AccResourceFromAccount(acc)
	keeper.vmKeeper.SetValue(ctx, &vm.VMAccessPath{
		Address: AddrToPathAddr(acc.GetAddress()),
		Path:    GetResPath(),
	}, AccToBytes(accRes))
}

func (keeper VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	account := keeper.AccountKeeper.GetAccount(ctx, addr)

	// check if account maybe exists in vm storage.
	bz := keeper.vmKeeper.GetValue(ctx, &vm.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	})

	// if account exists, but only in vm.
	if bz != nil {
		accRes := BytesToAccRes(bz)
		realCoins := bytesToCoins(accRes.Balances)

		// load vm account from storage.
		// check if account exists in vm but not exists in our storage - if so, save account and return.
		// check if account has differences - balances, something else, and if so - save account and return.
		if account != nil {
			if !realCoins.IsEqual(account.GetCoins()) { // also check coins
				account.SetCoins(realCoins)

				keeper.SetAccount(ctx, account)
			}
		} else {
			// if account is not exists - so create it.
			account = keeper.NewAccountWithAddress(ctx, addr)
			account.SetCoins(realCoins)
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

// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (keeper VMAccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	keeper.RemoveAccount(ctx, acc)

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

	return nil, sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", addr)).Result()
}

/*
// RemoveAccount removes an account for the account mapper store.
// NOTE: this will cause supply invariant violation if called
func (ak AccountKeeper) RemoveAccount(ctx sdk.Context, acc exported.Account) {
	addr := acc.GetAddress()
	store := ctx.KVStore(ak.key)
	store.Delete(types.AddressStoreKey(addr))
}
*/
/*

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	accounts := []exported.Account{}
	appendAccount := func(acc exported.Account) (stop bool) {
		accounts = append(accounts, acc)
		return false
	}
	ak.IterateAccounts(ctx, appendAccount)
	return accounts
}
*/
/*
func NewAccountKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramstore subspace.Subspace, proto func() exported.Account,
) AccountKeeper {

	return AccountKeeper{
		key:           key,
		proto:         proto,
		cdc:           cdc,
		paramSubspace: paramstore.WithKeyTable(types.ParamKeyTable()),
	}
}
*/
