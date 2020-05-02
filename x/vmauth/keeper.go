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

// Get account from VM storage.
// If no account found, second return parameter is false.
func (keeper VMAccountKeeper) getVMAccount(ctx sdk.Context, address sdk.AccAddress) (AccountResource, bool) {
	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(address),
		Path:    GetResPath(),
	}

	val := keeper.vmKeeper.GetValue(ctx, accessPath)
	if val == nil {
		return AccountResource{}, false
	}

	return BytesToAccRes(val), true
}

// Set account for VM.
func (keeper VMAccountKeeper) setVMAccount(ctx sdk.Context, address sdk.AccAddress, vmAccount AccountResource) {
	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.Bech32ToLibra(address),
		Path:    GetResPath(),
	}

	keeper.vmKeeper.SetValue(ctx, accessPath, AccResToBytes(vmAccount))
}

// Save new VM account.
func (keeper VMAccountKeeper) saveNewVMAccount(ctx sdk.Context, address sdk.AccAddress, vmAccount AccountResource, eventHandleGen EventHandleGenerator) {
	vmAddr := common_vm.Bech32ToLibra(address)
	accessPath := &vm_grpc.VMAccessPath{
		Address: vmAddr,
		Path:    GetResPath(),
	}

	bz := AccResToBytes(vmAccount)
	keeper.vmKeeper.SetValue(ctx, accessPath, bz)
	keeper.vmKeeper.SetValue(ctx, &vm_grpc.VMAccessPath{
		Address: vmAddr,
		Path:    GetEHPath(),
	}, EventHandlerGenToBytes(eventHandleGen))
}

// Save balances in VM keeper.
func (keeper VMAccountKeeper) saveBalances(ctx sdk.Context, balances Balances) {
	for _, balance := range balances {
		keeper.vmKeeper.SetValue(ctx, balance.accessPath, BalanceToBytes(balance.balance))
	}
}

// Load balances from VM storage.
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

// Set account in storage.
func (keeper VMAccountKeeper) SetAccount(ctx sdk.Context, acc exported.Account) {
	addr := acc.GetAddress()
	source, isExists := keeper.getVMAccount(ctx, addr)

	if isExists {
		mergedVMAccount := MergeVMAccountEvents(acc, source)
		keeper.setVMAccount(ctx, addr, mergedVMAccount)
	} else {
		vmAccount, eventHandleGen := CreateVMAccount(acc)
		keeper.saveNewVMAccount(ctx, addr, vmAccount, eventHandleGen)
	}

	// update balances extracted from coins
	balances := coinsToBalances(acc)
	keeper.saveBalances(ctx, balances)

	keeper.AccountKeeper.SetAccount(ctx, acc)
}

// Get account from storage.
func (keeper VMAccountKeeper) GetAccount(ctx sdk.Context, addr sdk.AccAddress) exported.Account {
	account := keeper.AccountKeeper.GetAccount(ctx, addr)

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
		Address: common_vm.Bech32ToLibra(acc.GetAddress()),
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
