// +build unit

package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Generate VM arguments.
func newArgInputs() []types.ScriptArg {
	args := make([]types.ScriptArg, 0)

	// Bool: true
	{
		tag, _ := vm_client.NewBoolScriptArg("true")
		args = append(args, tag)
	}
	// Bool: false
	{
		tag, _ := vm_client.NewBoolScriptArg("false")
		args = append(args, tag)
	}
	// Vector
	{
		tag, _ := vm_client.NewVectorScriptArg("0x010203040506070809AABBCCDDEEFF")
		args = append(args, tag)
	}
	// Address
	{
		addr := sdk.AccAddress(randomValue(common_vm.VMAddressLength))
		tag, _ := vm_client.NewAddressScriptArg(addr.String())
		args = append(args, tag)
	}
	// U8
	{
		tag, _ := vm_client.NewU8ScriptArg("128")
		args = append(args, tag)
	}
	// U64
	{
		tag, _ := vm_client.NewU64ScriptArg("1000000")
		args = append(args, tag)
	}
	// U128
	{
		tag, _ := vm_client.NewU128ScriptArg("100000000000000000000000000000")
		args = append(args, tag)
	}

	return args
}

// Get free gas calculations.
func TestVMKeeper_GetFreeGas(t *testing.T) {
	t.Parallel()

	var gasLimit uint64 = 1000

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	ctx := sdk.NewContext(mstore, abci.Header{ChainID: "dn-testnet-vm-keeper-test"}, false, log.NewNopLogger())
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))

	freeGas := GetFreeGas(ctx)

	require.Equal(t, gasLimit, freeGas)

	var gasToSpend uint64 = 900
	restGas := gasLimit - gasToSpend

	ctx.GasMeter().ConsumeGas(gasToSpend, "test spend")

	freeGas = GetFreeGas(ctx)
	require.Equal(t, restGas, freeGas)

	ctx.GasMeter().ConsumeGas(restGas, "test spend to zero")
	freeGas = GetFreeGas(ctx)
	require.EqualValues(t, 0, freeGas)
}

// Check creation of new contract instance.
func TestVMKeeper_NewContract(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress(randomValue(common_vm.VMAddressLength))
	code := randomValue(1024)
	argInputs := newArgInputs()
	maxGas := uint64(1000000)

	contractModule := NewDeployContract(addr, maxGas, code)
	require.Equal(t, common_vm.Bech32ToLibra(addr), contractModule.Address)
	require.Equal(t, maxGas, contractModule.MaxGasAmount)
	require.Equal(t, uint64(types.VmGasPrice), contractModule.GasUnitPrice)
	require.Equal(t, code, contractModule.Code)

	contractScript, err := NewExecuteContract(addr, maxGas, code, argInputs)
	require.NoError(t, err)
	require.Equal(t, common_vm.Bech32ToLibra(addr), contractScript.Address)
	require.Equal(t, maxGas, contractScript.MaxGasAmount)
	require.Equal(t, uint64(types.VmGasPrice), contractScript.GasUnitPrice)
	require.Equal(t, code, contractScript.Code)
	require.Equal(t, len(argInputs), len(contractScript.Args))
	for i, contractArg := range contractScript.Args {
		require.Equal(t, argInputs[i].Type, contractArg.Type)
		require.Equal(t, argInputs[i].Value, contractArg.Value)
	}
}

// Create new deploy request.
func TestVMKeeper_NewDeployRequest(t *testing.T) {
	t.Parallel()

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	code := randomValue(1024)

	var gasLimit uint64 = 10000000

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	ctx := sdk.NewContext(mstore, abci.Header{ChainID: "dn-testnet-vm-keeper-test"}, false, log.NewNopLogger())
	ctx = ctx.WithGasMeter(sdk.NewGasMeter(gasLimit))

	msg := types.MsgDeployModule{
		Signer: addr,
		Module: code,
	}

	req, err := NewDeployRequest(ctx, msg)
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, common_vm.Bech32ToLibra(addr), req.Address)
	require.EqualValues(t, gasLimit, req.MaxGasAmount)
	require.EqualValues(t, types.VmGasPrice, req.GasUnitPrice)
	require.EqualValues(t, code, req.Code)
}
