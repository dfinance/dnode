// +build unit

package keeper

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

type argInput struct {
	vmType  vm_grpc.VMTypeTag
	vmValue []byte
	value   string
}

type argInputs []argInput

func (list argInputs) getScriptArgs() []types.ScriptArg {
	args := make([]types.ScriptArg, 0, len(list))
	for _, input := range list {
		args = append(args, types.ScriptArg{
			Type:  input.vmType,
			Value: input.value,
		})
	}

	return args
}

// Generate VM arguments.
func newArgInputs() argInputs {
	inputs := make(argInputs, 0)

	// Bool: true
	{
		value := "true"
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_Bool,
			vmValue: []byte(value),
			value:   value,
		})
	}
	// Bool: false
	{
		value := "false"
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_Bool,
			vmValue: []byte(value),
			value:   value,
		})
	}
	// Vector
	{
		vector := randomValue(32)
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_Vector,
			vmValue: vector,
			value:   hex.EncodeToString(vector),
		})
	}
	// Address
	{
		addr := sdk.AccAddress(randomValue(common_vm.VMAddressLength))
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_Address,
			vmValue: common_vm.Bech32ToLibra(addr),
			value:   addr.String(),
		})
	}
	// U8
	{
		value := "128"
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_U8,
			vmValue: []byte(value),
			value:   value,
		})
	}
	// U64
	{
		value := "1000000"
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_U64,
			vmValue: []byte(value),
			value:   value,
		})
	}
	// U128
	{
		value := "100000000000000000000000000000"
		inputs = append(inputs, argInput{
			vmType:  vm_grpc.VMTypeTag_U128,
			vmValue: []byte(value),
			value:   value,
		})
	}

	return inputs
}

// Get free gas calculations.
func TestGetFreeGas(t *testing.T) {
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
func TestNewContract(t *testing.T) {
	addr := sdk.AccAddress(randomValue(common_vm.VMAddressLength))
	code := randomValue(1024)
	argInputs := newArgInputs()
	maxGas := uint64(1000000)

	contractModule := NewDeployContract(addr, maxGas, code)
	require.Equal(t, common_vm.Bech32ToLibra(addr), contractModule.Address)
	require.Equal(t, maxGas, contractModule.MaxGasAmount)
	require.Equal(t, uint64(types.VmGasPrice), contractModule.GasUnitPrice)
	require.Equal(t, code, contractModule.Code)

	contractScript, err := NewExecuteContract(addr, maxGas, code, argInputs.getScriptArgs())
	require.NoError(t, err)
	require.Equal(t, common_vm.Bech32ToLibra(addr), contractScript.Address)
	require.Equal(t, maxGas, contractScript.MaxGasAmount)
	require.Equal(t, uint64(types.VmGasPrice), contractScript.GasUnitPrice)
	require.Equal(t, code, contractScript.Code)
	require.Equal(t, len(argInputs), len(contractScript.Args))
	for i, contractArg := range contractScript.Args {
		require.Equal(t, argInputs[i].vmType, contractArg.Type)
		require.Equal(t, argInputs[i].vmValue, contractArg.Value)
	}
}

// Create new deploy request.
func TestNewDeployRequest(t *testing.T) {
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
