// +build unit

// Cover empty fee and wrong denom fee with test.
// The rest of antehandler tests in x/auth/ante_test.go tests.
package core

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestTypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vmauth"
)

var (
	WrongFees = sdk.Coins{sdk.NewCoin("eth", sdk.NewInt(1))} // wrong fees denom (eth).
)

type testInput struct {
	cdc          *codec.Codec
	ctx          sdk.Context
	paramsKeeper params.Keeper
	supplyKeeper authTypes.SupplyKeeper
	vmStorage    vm.Keeper
	ccsStorage   ccstorage.Keeper
	accKeeper    vmauth.Keeper
}

// nolint:errcheck
func setupTestInput() testInput {
	input := testInput{
		cdc: codec.New(),
	}

	// register codec
	vestTypes.RegisterCodec(input.cdc)
	vmauth.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// create storage keys
	keyParams := sdk.NewKVStoreKey(params.StoreKey)
	vmKey := sdk.NewKVStoreKey(vm.StoreKey)
	ccsKey := sdk.NewKVStoreKey(ccstorage.StoreKey)
	accKey := sdk.NewKVStoreKey(authTypes.StoreKey)
	tkeyParams := sdk.NewTransientStoreKey(params.TStoreKey)

	// init in-memory DB
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(vmKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(ccsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(accKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	// create target and dependant keepers
	input.paramsKeeper = params.NewKeeper(input.cdc, keyParams, tkeyParams)
	input.vmStorage = vm.NewKeeper(
		input.cdc,
		vmKey,
		nil,
		nil,
		nil,
		ccstorage.RequestVMStoragePerms(),
	)
	input.ccsStorage = ccstorage.NewKeeper(
		input.cdc,
		ccsKey,
		input.paramsKeeper.Subspace(ccstorage.DefaultParamspace),
		input.vmStorage,
		vmauth.RequestCCStoragePerms(),
	)
	input.accKeeper = vmauth.NewKeeper(input.cdc, accKey, input.paramsKeeper.Subspace(auth.DefaultParamspace), input.ccsStorage, authTypes.ProtoBaseAccount)
	input.supplyKeeper = mock.NewDummySupplyKeeper(input.accKeeper.AccountKeeper)

	// create context
	input.ctx = sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id", Height: 1}, false, log.NewNopLogger())

	// init genesis and params
	input.ccsStorage.InitDefaultGenesis(input.ctx)
	input.accKeeper.SetParams(input.ctx, authTypes.DefaultParams())

	return input
}

// run the tx through the anteHandler and ensure it fails with the given code.
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool, expectedErr error) {
	_, err := anteHandler(ctx, tx, simulate)
	require.Error(t, err)
}

// run the tx through the anteHandler and ensure its valid.
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, err := anteHandler(ctx, tx, simulate)
	require.NoError(t, err)
}

// nolint:errcheck
// Test when no fees provided in transaction.
func TestAnteHandler_WrongZeroFee(t *testing.T) {
	t.Parallel()

	input := setupTestInput()

	priv, _, addr := vestTypes.KeyTestPubAddr()
	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr)

	acc.SetCoins(DefaultFees)
	input.accKeeper.SetAccount(input.ctx, acc)

	msg := vestTypes.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000}

	msgs := []sdk.Msg{msg}

	// test empty fees
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := authTypes.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.accKeeper, input.supplyKeeper, auth.DefaultSigVerificationGasConsumer)
	checkInvalidTx(t, ah, input.ctx, tx, true, ErrFeeRequired)
}

// nolint:errcheck
// Test when wrong denom is provided in transaction.
func TestAnteHandler_WrongFeeDenom(t *testing.T) {
	t.Parallel()

	input := setupTestInput()

	priv, _, addr := vestTypes.KeyTestPubAddr()
	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr)
	acc.SetCoins(DefaultFees)

	input.accKeeper.SetAccount(input.ctx, acc)

	msg := vestTypes.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000, Amount: WrongFees}

	msgs := []sdk.Msg{msg}

	// test wrong fees denom.
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := authTypes.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.accKeeper, input.supplyKeeper, auth.DefaultSigVerificationGasConsumer)
	checkInvalidTx(t, ah, input.ctx, tx, true, ErrWrongFeeDenom)
}

// nolint:errcheck
// Test for correct transaction with correct fees.
func TestAnteHandler_CorrectDenomFees(t *testing.T) {
	t.Parallel()

	input := setupTestInput()

	priv, _, addr := vestTypes.KeyTestPubAddr()
	acc := input.accKeeper.NewAccountWithAddress(input.ctx, addr)

	acc.SetCoins(DefaultFees)

	input.accKeeper.SetAccount(input.ctx, acc)
	msg := vestTypes.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000, Amount: DefaultFees}

	msgs := []sdk.Msg{msg}

	// test correct fees denom.
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := authTypes.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.accKeeper, input.supplyKeeper, auth.DefaultSigVerificationGasConsumer)
	checkValidTx(t, ah, input.ctx, tx, true)
}
