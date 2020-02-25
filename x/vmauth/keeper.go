package vmauth

import (
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/vm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	codec "github.com/tendermint/go-amino"
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

	if account != nil {
		bz := keeper.vmKeeper.GetValue(ctx, &vm.VMAccessPath{
			Address: AddrToPathAddr(addr),
			Path:    GetResPath(),
		})

		// load vm account from storage.
		// check if account exists in vm but not exists in our storage - if so, save account and return.
		// check if account has differences - balances, something else, and if so - save account and return.
		accRes := BytesToAccRes(bz)
		realCoins := bytesToCoins(accRes.Coins)
		if accRes.Sequence != account.GetSequence() || !realCoins.IsEqual(account.GetCoins()) { // also check coins
			account.SetCoins(realCoins)
			account.SetSequence(accRes.Sequence)

			keeper.SetAccount(ctx, account)
		}
	}

	return account
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (keeper VMAccountKeeper) GetAllAccounts(ctx sdk.Context) []exported.Account {
	accounts := keeper.AccountKeeper.GetAllAccounts(ctx)

	// get all accounts from vm storage, compare changes, save, if account exists only in vm storage - create and save.
	for i := range accounts {
		bz := keeper.vmKeeper.GetValue(ctx, &vm.VMAccessPath{
			Address: AddrToPathAddr(accounts[i].GetAddress()),
			Path:    GetResPath(),
		})

		accRes := BytesToAccRes(bz)
		realCoins := bytesToCoins(accRes.Coins)
		if accRes.Sequence != accounts[i].GetSequence() || !realCoins.IsEqual(accounts[i].GetCoins()) { // also check coins
			accounts[i].SetCoins(realCoins)
			accounts[i].SetSequence(accRes.Sequence)

			keeper.SetAccount(ctx, accounts[i])
		}
	}

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
