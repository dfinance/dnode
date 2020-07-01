package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"

	"github.com/dfinance/dnode/x/cc_storage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// GetResStdCurrencyInfo returns VM currencyInfo for stdlib currencies (non-token).
func (k Keeper) GetResStdCurrencyInfo(ctx sdk.Context, denom string) (types.ResCurrencyInfo, error) {
	path, err := k.GetCurrencyInfoPath(ctx, denom)
	if err != nil {
		return types.ResCurrencyInfo{}, sdkErrors.Wrapf(types.ErrWrongDenom, err.Error())
	}

	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    path,
	}

	if !k.vmKeeper.HasValue(ctx, accessPath) {
		return types.ResCurrencyInfo{}, sdkErrors.Wrapf(types.ErrInternal, "currencyInfo for %q: nof found in VM storage", denom)
	}

	currencyInfo := types.ResCurrencyInfo{}
	bz := k.vmKeeper.GetValue(ctx, accessPath)
	if err := lcs.Unmarshal(bz, &currencyInfo); err != nil {
		return types.ResCurrencyInfo{}, sdkErrors.Wrapf(types.ErrInternal, "currencyInfo for %q: lcs unmarshal: %v", denom, err)
	}

	return currencyInfo, nil
}

// storeResStdCurrencyInfo sets currencyInfo to the VM storage.
func (k Keeper) storeResStdCurrencyInfo(ctx sdk.Context, currency types.Currency) {
	currencyInfo, err := types.NewResCurrencyInfo(currency, common_vm.StdLibAddress)
	if err != nil {
		panic(fmt.Errorf("currency %q: %v", currency.Denom, err))
	}

	path, err := k.GetCurrencyInfoPath(ctx, currency.Denom)
	if err != nil {
		panic(fmt.Errorf("currency %q: %v", currency.Denom, err))
	}

	bz, err := lcs.Marshal(currencyInfo)
	if err != nil {
		panic(fmt.Errorf("currency %q: lcs marshal: %v", currency.Denom, err))
	}

	accessPath := &vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    path,
	}

	k.vmKeeper.SetValue(ctx, accessPath, bz)
}
