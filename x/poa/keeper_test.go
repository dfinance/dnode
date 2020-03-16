package poa

import (
	"testing"

	poatypes "github.com/WingsDao/wings-blockchain/x/poa/types"
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
)

const (
	ethAddress1 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE86B"
	ethAddress2 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE87B"
	ethAddress3 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE88B"
	ethAddress4 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE89B"
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
	keySupply  *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	keyPoa     *sdk.KVStoreKey

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
		//keyCC:      sdk.NewKVStoreKey("cc"),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyParams:  sdk.NewKVStoreKey("params"),
		tkeyParams: sdk.NewTransientStoreKey("transient_params"),
		keyPoa:     sdk.NewKVStoreKey("poa"),
		//keyMS:      sdk.NewKVStoreKey("multisig"),
	}

	RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	bank.RegisterCodec(input.cdc)
	staking.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)
	//multisig.RegisterCodec(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	//mstore.MountStoreWithDB(input.keyCC, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyPoa, sdk.StoreTypeIAVL, db)
	//mstore.MountStoreWithDB(input.keyMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	if err := mstore.LoadLatestVersion(); err != nil {
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

	// Initializing POA module
	input.target = NewKeeper(
		input.keyPoa,
		input.cdc,
		input.paramsKeeper.Subspace("poa"),
	)

	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	// input.accountKeeper.SetParams(input.ctx, auth.DefaultParams())
	// input.bankKeeper.SetSendEnabled(input.ctx, true)

	return input
}

func checkError(t *testing.T, expectedErr, receivedErr sdk.Error) {
	require.Equal(t, expectedErr.Codespace(), receivedErr.Codespace(), "Codespace")
	require.Equal(t, expectedErr.Code(), receivedErr.Code(), "code")
}

// func TestKeeper_GetCDC(t *testing.T) {
// 	t.Parallel()
//
// 	input := setupTestInput(t)
// 	target := input.target
//
// 	cdc := target.GetCDC()
// 	require.Equal(t, input.cdc, cdc)
// }

func TestKeeper_AddValidator(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	target.AddValidator(ctx, addr, ethAddress1)
	require.True(t, target.HasValidator(ctx, addr))
}

func TestKeeper_GetValidators(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	require.Equal(t, 0, len(target.GetValidators(ctx)))
	target.AddValidator(ctx, addr, ethAddress1)
	require.True(t, target.HasValidator(ctx, addr))
	require.Equal(t, 1, len(target.GetValidators(ctx)))
}

func TestKeeper_GetValidatorAmount(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	require.Equal(t, uint16(0), target.GetValidatorAmount(ctx))

	require.Equal(t, 0, len(target.GetValidators(ctx)))
	target.AddValidator(ctx, addr, ethAddress1)
	require.Equal(t, uint16(1), target.GetValidatorAmount(ctx))
}

func TestKeeper_GetValidator(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	{
		validator := target.GetValidator(ctx, addr)
		require.Nil(t, validator.Address)
		require.Zero(t, validator.EthAddress)
	}
	target.AddValidator(ctx, addr, ethAddress1)
	require.True(t, target.HasValidator(ctx, addr))
	{
		validator := target.GetValidator(ctx, addr)
		require.NotNil(t, validator.Address)
		require.NotZero(t, validator.EthAddress)
	}
}

func TestKeeper_HasValidator(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	require.False(t, target.HasValidator(ctx, addr))
	target.AddValidator(ctx, addr, ethAddress1)
	require.True(t, target.HasValidator(ctx, addr))
	require.Equal(t, 1, len(target.GetValidators(ctx)))
}

func TestKeeper_GetEnoughConfirmations(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	require.Equal(t, uint16(1), target.GetEnoughConfirmations(ctx))
	target.AddValidator(ctx, addr, ethAddress1)
	target.AddValidator(ctx, addr, ethAddress2)
	target.AddValidator(ctx, addr, ethAddress3)
	target.AddValidator(ctx, addr, ethAddress4)
	require.Equal(t, uint16(3), target.GetEnoughConfirmations(ctx))
}

func TestKeeper_RemoveValidator(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr := sdk.AccAddress([]byte("addr1"))

	target.AddValidator(ctx, addr, ethAddress1)
	target.AddValidator(ctx, addr, ethAddress2)
	require.True(t, target.HasValidator(ctx, addr))
	require.Equal(t, uint16(2), target.GetValidatorAmount(ctx))
	target.RemoveValidator(ctx, addr)
	require.Equal(t, uint16(1), target.GetValidatorAmount(ctx))
	target.RemoveValidator(ctx, addr)
	require.Equal(t, uint16(0), target.GetValidatorAmount(ctx))
	require.False(t, target.HasValidator(ctx, addr))
}

func TestKeeper_ReplaceValidator(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target
	addr1 := sdk.AccAddress([]byte("addr1"))
	addr2 := sdk.AccAddress([]byte("addr2"))

	target.AddValidator(ctx, addr1, ethAddress1)
	require.True(t, target.HasValidator(ctx, addr1))
	require.Equal(t, uint16(1), target.GetValidatorAmount(ctx))
	target.ReplaceValidator(ctx, addr1, addr2, ethAddress2)
	require.Equal(t, uint16(1), target.GetValidatorAmount(ctx))
	v := target.GetValidator(ctx, addr2)
	require.Equal(t, addr2.String(), v.Address.String())
	require.Equal(t, ethAddress2, v.EthAddress)
}

func TestKeeper_ExportGenesis(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target

	addr, ethAddr := sdk.AccAddress([]byte("addr1")), ethAddress1

	initGenesis := poatypes.GenesisState{
		Parameters: poatypes.DefaultParams(),
		PoAValidators: poatypes.Validators{
			poatypes.Validator{
				Address:    addr,
				EthAddress: ethAddr,
			},
		},
	}

	target.InitGenesis(ctx, initGenesis)
	exportGenesis := target.ExportGenesis(ctx)

	require.True(t, initGenesis.Parameters.Equal(exportGenesis.Parameters))
	require.ElementsMatch(t, initGenesis.PoAValidators, exportGenesis.PoAValidators)
}

func TestModule_ValidateGenesis(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	app := NewAppMsModule(input.target)
	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	genesis := poatypes.GenesisState{
		Parameters:    poatypes.DefaultParams(),
		PoAValidators: poatypes.Validators{},
	}

	// check minValidators error
	if genesis.Parameters.MinValidators > 1 {
		err := app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis))
		checkError(t, poatypes.ErrNotEnoungValidators(0, 0), err.(sdk.Error))
	}

	// check OK
	{
		for i := uint16(0); i < genesis.Parameters.MaxValidators; i++ {
			genesis.PoAValidators = append(genesis.PoAValidators, poatypes.Validator{
				Address:    sdkAddress,
				EthAddress: ethAddress1,
			})
		}
		require.Nil(t, app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis)))
	}

	// check maxValidators error
	{
		genesis.PoAValidators = append(genesis.PoAValidators, poatypes.Validator{
			Address:    sdkAddress,
			EthAddress: ethAddress1,
		})
		err := app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis))
		checkError(t, poatypes.ErrMaxValidatorsReached(0), err.(sdk.Error))
	}

	// check params validation
	if poatypes.DefaultMinValidators > 1 {
		genesis := poatypes.GenesisState{
			Parameters:    poatypes.Params{
				MaxValidators: poatypes.DefaultMinValidators - 1,
				MinValidators: 0,
			},
			PoAValidators: poatypes.Validators{},
		}
		require.NotNil(t, app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis)))

		genesis.Parameters.MinValidators = poatypes.DefaultMinValidators
		genesis.Parameters.MaxValidators = poatypes.DefaultMaxValidators + 1
		require.NotNil(t, app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis)))
	}
}
