package currencies

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/WingsDao/wings-blockchain/x/currencies/msgs"
	"github.com/WingsDao/wings-blockchain/x/currencies/types"
	"github.com/WingsDao/wings-blockchain/x/multisig"
)

const (
	symbol = "testcoin"
	issue1 = "issue1"
	issue2 = "issue2"
)

var (
	maccPerms map[string][]string = map[string][]string{
		auth.FeeCollectorName: nil,
	}
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyCC      *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	keyPoa     *sdk.KVStoreKey
	keyMS      *sdk.KVStoreKey

	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	paramsKeeper  params.Keeper

	target Keeper
}

func (ti *testInput) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func setupTestInput(t *testing.T) testInput {
	input := testInput{
		cdc:        codec.New(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyCC:      sdk.NewKVStoreKey("cc"),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyParams:  sdk.NewKVStoreKey("params"),
		tkeyParams: sdk.NewTransientStoreKey("transient_params"),
		keyPoa:     sdk.NewKVStoreKey("poa"),
		keyMS:      sdk.NewKVStoreKey("multisig"),
	}

	RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	bank.RegisterCodec(input.cdc)
	staking.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)
	multisig.RegisterCodec(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCC, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyPoa, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	// The ParamsKeeper handles parameter storage for the application
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams, params.DefaultCodespace)

	// The AccountKeeper handles address -> account lookups
	input.accountKeeper = auth.NewAccountKeeper(
		input.cdc,
		input.keyAccount,
		input.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	input.bankKeeper = bank.NewBaseKeeper(
		input.accountKeeper,
		input.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		input.ModuleAccountAddrs(),
	)

	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, maccPerms)

	// Initializing currencies module
	input.target = NewKeeper(
		input.bankKeeper,
		input.keyCC,
		input.cdc,
	)

	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	// input.accountKeeper.SetParams(input.ctx, auth.DefaultParams())
	// input.bankKeeper.SetSendEnabled(input.ctx, true)

	return input
}

// func TestKeeper_GetCDC(t *testing.T) {
// 	t.Parallel()
// 	input := setupTestInput(t)
//
// 	require.NotNil(t, input.target.GetCDC())
// }

func TestKeeper_IssueCurrency(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	input.accountKeeper.SetAccount(ctx, acc)

	amount := sdk.NewInt(10)
	require.NoError(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 2, addr, issue2))
	require.True(t, target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).Equal(amount))
}

func TestKeeper_DestroyCurrency(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	recipient := sdk.AccAddress([]byte("addr2"))
	input.accountKeeper.SetAccount(ctx, acc)
	amount := sdk.NewInt(10)

	// destroy unknown currency
	require.Error(t, target.DestroyCurrency(ctx, ctx.ChainID(), symbol, recipient.String(), amount, addr))

	// issue currency
	require.NoError(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 2, addr, issue2))
	require.True(t, target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).Equal(amount))

	// destroy currency
	require.NoError(t, target.DestroyCurrency(ctx, ctx.ChainID(), symbol, recipient.String(), amount, addr))
}

func TestKeeper_GetDestroy(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	recipient := sdk.AccAddress([]byte("addr2"))
	input.accountKeeper.SetAccount(ctx, acc)
	amount := sdk.NewInt(10)

	// destroy unknown currency
	require.Error(t, target.DestroyCurrency(ctx, ctx.ChainID(), symbol, recipient.String(), amount, addr))

	// issue currency
	require.NoError(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 2, addr, issue2))
	require.True(t, target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).Equal(amount))
	issue := target.GetIssue(ctx, issue1)
	require.Equal(t, addr.String(), issue.Recipient.String())
	require.Equal(t, amount.String(), issue.Amount.String())
	require.Equal(t, symbol, issue.Symbol)

	// destroy currency
	require.NoError(t, target.DestroyCurrency(ctx, ctx.ChainID(), symbol, recipient.String(), amount, addr))

	destroy := target.GetDestroy(ctx, target.getLastID(ctx))
	require.Equal(t, target.getLastID(ctx).String(), destroy.ID.String())
	require.True(t, target.HasDestroy(ctx, destroy.ID))

}

func TestKeeper_GetCurrency(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	input.accountKeeper.SetAccount(ctx, acc)
	amount := sdk.NewInt(10)

	// issue currency
	require.NoError(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 0, addr, issue1))
	require.Error(t, target.IssueCurrency(ctx, symbol, amount, 2, addr, issue2))
	require.True(t, target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).Equal(amount))
	issue := target.GetIssue(ctx, issue1)
	require.Equal(t, addr.String(), issue.Recipient.String())
	require.Equal(t, amount.String(), issue.Amount.String())
	require.Equal(t, symbol, issue.Symbol)

	currency := target.GetCurrency(ctx, symbol)
	require.Equal(t, symbol, currency.Symbol)
}

func TestKeeper(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	input.accountKeeper.SetAccount(ctx, acc)

	require.True(t, target.coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))
	require.False(t, target.hasIssue(ctx, "test"))

	bigInt, ok := new(big.Int).SetString("1000000000000000000000000000000000000000000000", 10)
	if !ok {
		t.Fatal("Too big!")
	}
	issueMsg := msgs.MsgIssueCurrency{
		Symbol:    "tst",
		Amount:    sdk.NewIntFromBigInt(bigInt),
		Decimals:  0,
		Recipient: addr,
	}
	require.Nil(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue1"))
	require.IsType(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue1"), types.ErrExistsIssue("issue1"))
	require.Nil(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue2"))
}

func Test_IssueHugeAmount(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))
	acc := input.accountKeeper.NewAccountWithAddress(ctx, addr)
	input.accountKeeper.SetAccount(ctx, acc)

	require.True(t, target.coinKeeper.GetCoins(ctx, addr).IsEqual(sdk.NewCoins()))
	require.False(t, target.hasIssue(ctx, issue1))

	bigInt, ok := new(big.Int).SetString("1000000000000000000000000000000000000000000000", 10)
	if !ok {
		t.Fatal("Too big!")
	}
	issueMsg := msgs.MsgIssueCurrency{
		Symbol:    symbol,
		Amount:    sdk.NewIntFromBigInt(bigInt),
		Decimals:  0,
		Recipient: addr,
	}
	require.Nil(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue1"))
	require.Equal(t, bigInt.String(), target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).BigInt().String())

	require.IsType(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue1"), types.ErrExistsIssue("issue1"))
	require.Nil(t, target.IssueCurrency(ctx, issueMsg.Symbol, issueMsg.Amount, issueMsg.Decimals, issueMsg.Recipient, "issue2"))
	require.Equal(t, big.NewInt(0).Add(bigInt, bigInt).String(), target.coinKeeper.GetCoins(ctx, addr).AmountOf(symbol).BigInt().String())
}
