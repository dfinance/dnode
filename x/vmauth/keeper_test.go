package vmauth

import (
	"github.com/WingsDao/wings-blockchain/x/vm"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestVMAccountKeeper_SetAccount(t *testing.T) {
	input := newTestInput(t)

	// Check just set.
	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	// Check set with coins.
	coin := types.NewCoin("wings", types.NewInt(1))
	acc = auth.NewBaseAccountWithAddress(addr)
	acc.SetCoins(types.Coins{coin})
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter = input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)
}

func TestVMAccountKeeper_GetAccount(t *testing.T) {
	input := newTestInput(t)

	// Check just get with vm storage.
	coin := types.NewCoin("wings", types.NewInt(1))

	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	acc.SetCoins(types.Coins{coin})
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	key := &vm.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	}

	accData := input.vmStorage.GetValue(input.ctx, key)
	accRes := BytesToAccRes(accData)

	require.Len(t, acc.GetCoins(), len(accRes.Balances))

	accData = input.vmStorage.GetValue(input.ctx, key)
	accRes = BytesToAccRes(accData)

	require.Len(t, acc.Coins, len(accRes.Balances))

	for i, coin := range getter.GetCoins() {
		require.EqualValues(t, coin.Denom, accRes.Balances[i].Denom)
		require.EqualValues(t, coin.Amount.String(), accRes.Balances[i].Value.String())
	}

	// Check get if there is write in vm storage, but not in cosmos storage.
	vmAcc := auth.NewBaseAccountWithAddress(types.AccAddress("vm"))
	vmAcc.SetCoins(types.Coins{coin})
	vmKey := &vm.VMAccessPath{
		Address: AddrToPathAddr(vmAcc.GetAddress()),
		Path:    GetResPath(),
	}

	bz := AccResToBytes(AccResFromAccount(&vmAcc))

	input.vmStorage.SetValue(input.ctx, vmKey, bz)

	getter = input.accountKeeper.GetAccount(input.ctx, vmAcc.GetAddress())
	require.EqualValues(t, &vmAcc, getter)

	// Returns nil.
	getter = input.accountKeeper.GetAccount(input.ctx, types.AccAddress("nil"))
	require.Nil(t, getter)
}

func TestVMAccountKeeper_RemoveAccount(t *testing.T) {
	input := newTestInput(t)
	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	input.accountKeeper.RemoveAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.Nil(t, getter)

	key := &vm.VMAccessPath{
		Address: AddrToPathAddr(addr),
		Path:    GetResPath(),
	}

	bz := input.vmStorage.GetValue(input.ctx, key)
	require.Empty(t, bz)
}

func TestGetSignerAcc(t *testing.T) {
	input := newTestInput(t)
	addr := types.AccAddress("tmp")
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter, res := GetSignerAcc(input.ctx, input.accountKeeper, addr)
	require.EqualValues(t, &acc, getter)
	require.EqualValues(t, res, types.Result{})

	getter, res = GetSignerAcc(input.ctx, input.accountKeeper, types.AccAddress("bmp"))
	require.Nil(t, getter)
	require.NotNil(t, res)
}
