// Implements account keeper with vm storage inside to allow work with accounts from VM.
package vmauth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
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

// Set account in storage.
func (keeper VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	keeper.AccountKeeper.SetAccount(ctx, acc)
	// check if account exists in vm
	accessPath := &vm_grpc.VMAccessPath{
		Address: acc.GetAddress(),
		Path:    GetResPath(),
	}

	vmBz := keeper.vmKeeper.GetValue(ctx, accessPath)
	if vmBz != nil {
		// get account from vm and copy event data
		// now store account to vm storage
		source := BytesToAccRes(vmBz)
		accRes := AccResFromAccount(acc, &source)
		keeper.vmKeeper.SetValue(ctx, accessPath, AccResToBytes(accRes))
	} else {
		// just create new account
		accRes := AccResFromAccount(acc, nil)
		keeper.vmKeeper.SetValue(ctx, accessPath, AccResToBytes(accRes))
	}
}

// Get account from storage.
func (keeper VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	account := keeper.AccountKeeper.GetAccount(ctx, addr)

	// check if account maybe exists in vm storage.
	bz := keeper.vmKeeper.GetValue(ctx, &vm_grpc.VMAccessPath{
		Address: addr,
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
					panic(err) // should never happen
				}

				keeper.SetAccount(ctx, account)
			}
		} else {
			// if account is not exists - so create it.
			account = keeper.NewAccountWithAddress(ctx, addr)
			if err := account.SetCoins(realCoins); err != nil {
				panic(err) // should never happen
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
	keeper.vmKeeper.DelValue(ctx, &vm_grpc.VMAccessPath{
		Address: acc.GetAddress(),
		Path:    GetResPath(),
	})
}

// GetSignerAcc returns an account for a given address that is expected to sign
// a transaction.
func GetSignerAcc(ctx sdk.Context, ak VMAccountKeeper, addr sdk.AccAddress) (exported.Account, error) {
	if acc := ak.GetAccount(ctx, addr); acc != nil {
		return acc, nil
	}

	return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q does not exist", addr)
}
