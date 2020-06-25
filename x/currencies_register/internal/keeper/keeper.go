package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// Keeper for registered currency.
type Keeper struct {
	cdc      *amino.Codec // Amino codec.
	storeKey sdk.StoreKey // Store key.

	vmStorage common_vm.VMStorage // virtual machine storage.
}

// Create new keeper.
func NewKeeper(cdc *amino.Codec, storeKey sdk.StoreKey, vmStorage common_vm.VMStorage) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		vmStorage: vmStorage,
	}
}

// Add currency info.
func (keeper Keeper) AddCurrencyInfo(ctx sdk.Context, denom string, decimals uint8, isToken bool, owner []byte, totalSupply sdk.Int, path []byte) error {
	store := ctx.KVStore(keeper.storeKey)
	keyPath := types.GetCurrencyPathKey(denom)

	// Check path is empty.
	if len(path) == 0 {
		return types.ErrInvalidPath
	}

	// Check currency already exists.
	if store.Has(keyPath) {
		return sdkErrors.Wrap(types.ErrExists, fmt.Sprintf("denom %q", denom))
	}

	// Save currency path
	currencyPath := types.NewCurrencyPath(path)
	bz, err := keeper.cdc.MarshalBinaryBare(currencyPath)
	if err != nil {
		return sdkErrors.Wrap(types.ErrInternal, err.Error())
	}
	store.Set(keyPath, bz)

	// Save to vm storage under owner account.
	currencyInfo, err := types.NewCurrencyInfo([]byte(denom), decimals, isToken, owner, totalSupply)
	if err != nil {
		return sdkErrors.Wrap(types.ErrWrongCurrencyInfo, err.Error())
	}

	bz, err = lcs.Marshal(currencyInfo)
	if err != nil {
		return sdkErrors.Wrapf(types.ErrLcsMarshal, "currencyInfo marshal: %v", err)
	}

	accessPath := vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    path,
	}
	keeper.vmStorage.SetValue(ctx, &accessPath, bz)

	return nil
}

// Check if currency already exists.
func (keeper Keeper) ExistsCurrencyInfo(ctx sdk.Context, denom string) bool {
	store := ctx.KVStore(keeper.storeKey)
	keyPath := types.GetCurrencyPathKey(denom)

	return store.Has(keyPath)
}

// Get currency info.
func (keeper Keeper) GetCurrencyInfo(ctx sdk.Context, denom string) (types.CurrencyInfo, error) {
	store := ctx.KVStore(keeper.storeKey)
	keyPath := types.GetCurrencyPathKey(denom)

	// Return error if currency already registered.
	if !store.Has(keyPath) {
		return types.CurrencyInfo{}, sdkErrors.Wrap(types.ErrNotFound, fmt.Sprintf("not found info with denom %q", denom))
	}

	// load path
	var currencyPath types.CurrencyPath
	bz := store.Get(keyPath)
	if err := keeper.cdc.UnmarshalBinaryBare(bz, &currencyPath); err != nil {
		return types.CurrencyInfo{}, sdkErrors.Wrap(types.ErrInternal, "unmarshal CurrencyPath")
	}

	accessPath := vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    currencyPath.Path,
	}

	// load resource
	bz = keeper.vmStorage.GetValue(ctx, &accessPath)
	var currInfo types.CurrencyInfo
	err := lcs.Unmarshal(bz, &currInfo)
	if err != nil {
		return types.CurrencyInfo{}, sdkErrors.Wrap(types.ErrLcsUnmarshal, fmt.Sprintf("can't unmarshal currency , denom %q: %v", denom, err))
	}

	return currInfo, nil
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/" + types.ModuleName)
}
