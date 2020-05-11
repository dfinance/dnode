package keeper

import (
	"crypto/rand"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

// Generate random bytes.
func randomBytes(l int) []byte {
	bz := make([]byte, l)

	rand.Read(bz)
	return bz
}

// Add currency info.
func TestKeeper_AddCurrencyInfo(t *testing.T) {
	denom := "dfi"
	decimals := 18
	isToken := false
	owner := make([]byte, common_vm.VMAddressLength)
	path := randomBytes(32)
	totalSupply, isOk := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, isOk)

	input := GetTestInput(t)

	err := input.keeper.AddCurrencyInfo(input.ctx, denom, uint8(decimals), isToken, owner, totalSupply, path)
	require.NoError(t, err)

	accessPath := vm_grpc.VMAccessPath{
		Address: common_vm.ZeroAddress,
		Path:    path,
	}

	isExists := input.keeper.vmStorage.HasValue(input.ctx, &accessPath)
	require.True(t, isExists)

	var currInfo types.CurrencyInfo
	err = lcs.Unmarshal(input.keeper.vmStorage.GetValue(input.ctx, &accessPath), &currInfo)
	require.NoError(t, err)

	require.EqualValues(t, denom, currInfo.Denom)
	require.EqualValues(t, decimals, currInfo.Decimals)
	require.EqualValues(t, isToken, currInfo.IsToken)
	require.EqualValues(t, owner, currInfo.Owner)
	require.EqualValues(t, totalSupply.String(), currInfo.TotalSupply.String())
}

// Add currency info and returned errors.
func TestKeeper_AddCurrencyInfoErrors(t *testing.T) {
	denom := "dfi"
	anotherDenom := "eth"
	decimals := 18
	isToken := false
	owner := make([]byte, common_vm.VMAddressLength)
	path := randomBytes(32)
	totalSupply, isOk := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, isOk)

	input := GetTestInput(t)

	err := input.keeper.AddCurrencyInfo(input.ctx, denom, uint8(decimals), isToken, owner, totalSupply, path)
	require.NoError(t, err)

	// Check error when same denom added again.
	err = input.keeper.AddCurrencyInfo(input.ctx, denom, uint8(decimals), isToken, owner, totalSupply, path)
	require.Error(t, err)
	require.Equal(t, err.Error(), fmt.Sprintf("currency already exists: denom %q", denom))

	// Add currency with wrong owner address.
	err = input.keeper.AddCurrencyInfo(input.ctx, anotherDenom, uint8(decimals), isToken, make([]byte, 40), totalSupply, path)
	require.Error(t, err)
	require.Contains(t, err.Error(), "wrong length address:")
}

// Exists currency info.
func TestKeeper_ExistsCurrencyInfo(t *testing.T) {
	denom := "dfi"
	decimals := 18
	isToken := false
	owner := make([]byte, common_vm.VMAddressLength)
	path := randomBytes(32)
	totalSupply, isOk := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, isOk)

	input := GetTestInput(t)

	err := input.keeper.AddCurrencyInfo(input.ctx, denom, uint8(decimals), isToken, owner, totalSupply, path)
	require.NoError(t, err)

	isExists := input.keeper.ExistsCurrencyInfo(input.ctx, denom)
	require.True(t, isExists)
}

// Get currency info.
func TestKeeper_GetCurrencyInfo(t *testing.T) {
	denom := "dfi"
	decimals := 18
	isToken := false
	owner := make([]byte, common_vm.VMAddressLength)
	path := randomBytes(32)
	totalSupply, isOk := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, isOk)

	input := GetTestInput(t)

	err := input.keeper.AddCurrencyInfo(input.ctx, denom, uint8(decimals), isToken, owner, totalSupply, path)
	require.NoError(t, err)

	currInfo, err := input.keeper.GetCurrencyInfo(input.ctx, denom)
	require.NoError(t, err)

	require.EqualValues(t, denom, currInfo.Denom)
	require.EqualValues(t, decimals, currInfo.Decimals)
	require.EqualValues(t, isToken, currInfo.IsToken)
	require.EqualValues(t, owner, currInfo.Owner)
	require.EqualValues(t, totalSupply.String(), currInfo.TotalSupply.String())
}

// Test get with errors.
func TestKeeper_GetCurrencyInfoErrors(t *testing.T) {
	input := GetTestInput(t)
	denom := "dfi"
	path := randomBytes(32)

	// Get non exists currency.
	_, err := input.keeper.GetCurrencyInfo(input.ctx, denom)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("not found info with denom %q", denom))

	// Cant unmarshal.
	accessPath := vm_grpc.VMAccessPath{
		Address: common_vm.ZeroAddress,
		Path:    path,
	}

	store := input.ctx.KVStore(input.keeper.storeKey)
	keyPath := types.GetCurrencyPathKey(denom)

	currencyPath := types.NewCurrencyPath(path)
	bz, err := input.keeper.cdc.MarshalBinaryBare(currencyPath)
	require.NoError(t, err)

	store.Set(keyPath, bz)

	input.keeper.vmStorage.SetValue(input.ctx, &accessPath, make([]byte, 2))

	_, err = input.keeper.GetCurrencyInfo(input.ctx, denom)
	require.Error(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("can't unmarshal currency , denom %q", denom))
}
