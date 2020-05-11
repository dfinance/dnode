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

// Generate VM arguments.
func getArgs() []*vm_grpc.VMArgs {
	addr := sdk.AccAddress([]byte(randomValue(32)))
	args := make([]*vm_grpc.VMArgs, 8)

	args[0] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_Bool,
		Value: "true",
	}

	args[1] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_Bool,
		Value: "false",
	}

	args[2] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_U64,
		Value: "1000000",
	}

	args[3] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_ByteArray,
		Value: "0x" + hex.EncodeToString(randomValue(32)),
	}

	args[4] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_Address,
		Value: addr.String(),
	}

	args[5] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_Struct,
		Value: "0x" + hex.EncodeToString(randomValue(64)),
	}

	args[6] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_U8,
		Value: "128",
	}

	args[7] = &vm_grpc.VMArgs{
		Type:  vm_grpc.VMTypeTag_U128,
		Value: "100000000000000000000000000000",
	}

	return args
}

// Get contract.
func getContract(addr sdk.AccAddress, contractType vm_grpc.ContractType, code []byte, maxGas uint64, args []*vm_grpc.VMArgs, t *testing.T) *vm_grpc.VMContract {
	contractModule, err := NewContract(addr, maxGas, code, contractType, args)
	if err != nil {
		t.Fatal(err)
	}

	return contractModule
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

// Check creationg of new contract instance.
func TestNewContract(t *testing.T) {
	addr := sdk.AccAddress(randomValue(32))
	code := randomValue(1024)
	args := getArgs()

	var maxGas uint64 = 1000000

	contractModule := getContract(addr, vm_grpc.ContractType_Module, code, maxGas, args, t)

	require.Equal(t, vm_grpc.ContractType_Module, contractModule.ContractType)
	require.Equal(t, maxGas, contractModule.MaxGasAmount)
	require.EqualValues(t, 1, contractModule.GasUnitPrice)

	for i, arg := range args {
		require.Equal(t, arg.Value, contractModule.Args[i].Value)
		require.Equal(t, arg.Type, contractModule.Args[i].Type)
	}

	// check script type
	contractScript := getContract(addr, vm_grpc.ContractType_Script, code, maxGas, args, t)

	require.Equal(t, vm_grpc.ContractType_Script, contractScript.ContractType)
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

	require.EqualValues(t, 0, req.Options)
	require.EqualValues(t, gasLimit, req.Contracts[0].MaxGasAmount)
	require.EqualValues(t, code, req.Contracts[0].Code)
	require.EqualValues(t, "0x"+hex.EncodeToString(common_vm.Bech32ToLibra(addr)), req.Contracts[0].Address)
	require.Equal(t, 1, len(req.Contracts))
}
