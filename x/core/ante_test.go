// Cover empty fee and wrong denom fee with test.
// The rest of antehandler tests in x/auth/ante_test.go tests.
package core

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	log "github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

var (
	WrongFees = sdk.Coins{sdk.NewCoin("eth", sdk.NewInt(1))} // wrong fees denom (eth).
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	ak  auth.AccountKeeper
	sk  types.SupplyKeeper
}

// nolint:errcheck
func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("subspace")
	tkeyParams := sdk.NewTransientStoreKey("transient_subspace")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	ps := subspace.NewSubspace(cdc, keyParams, tkeyParams, types.DefaultParamspace)
	ak := auth.NewAccountKeeper(cdc, authCapKey, ps, types.ProtoBaseAccount)
	sk := auth.NewDummySupplyKeeper(ak)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id", Height: 1}, false, log.NewNopLogger())

	ak.SetParams(ctx, types.DefaultParams())

	return testInput{cdc: cdc, ctx: ctx, ak: ak, sk: sk}
}

// run the tx through the anteHandler and ensure it fails with the given code.
func checkInvalidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool, code sdk.CodeType) {
	_, result, abort := anteHandler(ctx, tx, simulate)
	require.True(t, abort)

	require.Equal(t, code, result.Code, fmt.Sprintf("Expected %v, got %v", code, result))
	require.Equal(t, Codespace, result.Codespace)
}

// run the tx through the anteHandler and ensure its valid.
func checkValidTx(t *testing.T, anteHandler sdk.AnteHandler, ctx sdk.Context, tx sdk.Tx, simulate bool) {
	_, result, abort := anteHandler(ctx, tx, simulate)
	require.Equal(t, "", result.Log)
	require.False(t, abort)
	require.Equal(t, sdk.CodeOK, result.Code)
	require.True(t, result.IsOK())
}

// nolint:errcheck
// test when no fees provided in transaction.
func TestAnteHandlerWrongZeroFee(t *testing.T) {
	input := setupTestInput()

	priv, _, addr := types.KeyTestPubAddr()
	acc := input.ak.NewAccountWithAddress(input.ctx, addr)

	acc.SetCoins(DefaultFees)
	input.ak.SetAccount(input.ctx, acc)

	msg := types.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000}

	msgs := []sdk.Msg{msg}

	// test empty fees
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.ak, input.sk, auth.DefaultSigVerificationGasConsumer)
	checkInvalidTx(t, ah, input.ctx, tx, true, CodeFeeRequired)
}

// nolint:errcheck
// test when wrong denom provided in transaction.
func TestAnteHandlerWrongFeeDenom(t *testing.T) {
	input := setupTestInput()

	priv, _, addr := types.KeyTestPubAddr()
	acc := input.ak.NewAccountWithAddress(input.ctx, addr)
	acc.SetCoins(DefaultFees)

	input.ak.SetAccount(input.ctx, acc)

	msg := types.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000, Amount: WrongFees}

	msgs := []sdk.Msg{msg}

	// test wrong fees denom.
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.ak, input.sk, auth.DefaultSigVerificationGasConsumer)
	checkInvalidTx(t, ah, input.ctx, tx, true, CodeWrongFeeDenom)
}

// nolint:errcheck
// test for correct transaction with correct fees.
func TestCorrectDenomFees(t *testing.T) {
	input := setupTestInput()

	priv, _, addr := types.KeyTestPubAddr()
	acc := input.ak.NewAccountWithAddress(input.ctx, addr)

	acc.SetCoins(DefaultFees)

	input.ak.SetAccount(input.ctx, acc)
	msg := types.NewTestMsg(addr)
	fee := auth.StdFee{Gas: 10000, Amount: DefaultFees}

	msgs := []sdk.Msg{msg}

	// test correct fees denom.
	privs, accNums, seqs := []crypto.PrivKey{priv}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(input.ctx, msgs, privs, accNums, seqs, fee)

	ah := NewAnteHandler(input.ak, input.sk, auth.DefaultSigVerificationGasConsumer)
	checkValidTx(t, ah, input.ctx, tx, true)
}
