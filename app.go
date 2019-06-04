package app

import (
	"encoding/json"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
	"wings-blockchain/x/currencies"
	"wings-blockchain/x/poa"
	poaTypes "wings-blockchain/x/poa/types"
	msKeeper "wings-blockchain/x/multisig/keeper"
	"wings-blockchain/x/multisig"
)

const (
	appName = "wb"
)

type WbServiceApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyNS            *sdk.KVStoreKey
	keyCC		     *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey
	keyValidators    *sdk.KVStoreKey
	keyMS			 *sdk.KVStoreKey

	accountKeeper       auth.AccountKeeper
	bankKeeper          bank.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	paramsKeeper        params.Keeper
	currenciesKeeper    currencies.Keeper
	validatorsKeeper    poa.Keeper
	msKeeper			msKeeper.Keeper
}

// NewWbServiceApp is a constructor function for wings blockchain
func NewWbServiceApp(logger log.Logger, db dbm.DB) *WbServiceApp {

	// First define the top level codec that will be shared by the different modules
	cdc := MakeCodec()

	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc))

	// Here you initialize your application with the store keys it requires
	var app = &WbServiceApp{
		BaseApp: bApp,
		cdc:     cdc,

		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyCC:	  		  sdk.NewKVStoreKey("cc"),
		keyFeeCollection: sdk.NewKVStoreKey("fee_collection"),
		keyParams:        sdk.NewKVStoreKey("params"),
		tkeyParams:       sdk.NewTransientStoreKey("transient_params"),
		keyValidators:	  sdk.NewKVStoreKey("poa"),
		keyMS:            sdk.NewKVStoreKey("multisig"),
	}

	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// The FeeCollectionKeeper collects transaction fees and renders them to the fee distribution module
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(cdc, app.keyFeeCollection)

	// Initializing currencies module
	app.currenciesKeeper = currencies.NewKeeper(
		app.bankKeeper,
		app.keyCC,
		app.cdc,
	)

	// Initializing validators module
	app.validatorsKeeper = poa.NewKeeper(
		app.keyValidators,
		app.cdc,
		app.paramsKeeper.Subspace(poaTypes.DefaultParamspace),
	)

	// Initializing multisig router
	msRouter := msKeeper.NewRouter()
	msRouter.AddRoute("poa", poa.NewHandler(app.validatorsKeeper))

	// Initializing ms module
	app.msKeeper = msKeeper.NewKeeper(
		app.keyMS,
		app.cdc,
		msRouter,
	)

	// The AnteHandler handles signature verification and transaction pre-processing
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))

	// The app.Router is the main transaction router where each module registers its routes
	// Register the bank, currencies,  routes here
	app.Router().
		AddRoute("bank", bank.NewHandler(app.bankKeeper)).
		AddRoute("currencies", currencies.NewHandler(app.currenciesKeeper)).
		AddRoute("multisig", multisig.NewHandler(app.msKeeper, app.validatorsKeeper))
		//AddRoute("poa", poa.NewHandler(app.validatorsKeeper))

	// The app.QueryRouter is the main query router where each module registers its routes
	app.QueryRouter().
		AddRoute("acc", auth.NewQuerier(app.accountKeeper))

	// The initChainer handles translating the genesis.json file into initial state for the network
	app.SetInitChainer(app.initChainer)

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyCC,
		app.keyFeeCollection,
		app.keyParams,
		app.tkeyParams,
		app.keyValidators,
		app.keyMS,
	)

	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState struct {
	AuthData auth.GenesisState   `json:"auth"`
	BankData bank.GenesisState   `json:"bank"`
	Accounts []*auth.BaseAccount `json:"accounts"`
	PoAValidators []*poaTypes.Validator `json:"poa_validators"`
}

func (app *WbServiceApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	genesisState := new(GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, genesisState)
	if err != nil {
		panic(err)
	}

	app.validatorsKeeper.SetParams(ctx, poaTypes.DefaultParams())
	err = app.validatorsKeeper.InitGenesis(ctx, genesisState.PoAValidators)

	if err != nil {
		panic(err)
	}

	for _, acc := range genesisState.Accounts {
		acc.AccountNumber = app.accountKeeper.GetNextAccountNumber(ctx)
		app.accountKeeper.SetAccount(ctx, acc)
	}

	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)

	return abci.ResponseInitChain{}
}

// ExportAppStateAndValidators does the things
func (app *WbServiceApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
	ctx := app.NewContext(true, abci.Header{})
	accounts := []*auth.BaseAccount{}

	appendAccountsFn := func(acc auth.Account) bool {
		account := &auth.BaseAccount{
			Address: acc.GetAddress(),
			Coins:   acc.GetCoins(),
		}

		accounts = append(accounts, account)
		return false
	}

	app.accountKeeper.IterateAccounts(ctx, appendAccountsFn)

	genState := GenesisState{
		Accounts: accounts,
		AuthData: auth.DefaultGenesisState(),
		BankData: bank.DefaultGenesisState(),
	}

	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	return appState, validators, err
}

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	currencies.RegisterCodec(cdc)
	poa.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	multisig.RegisterCodec(cdc)
	return cdc
}
