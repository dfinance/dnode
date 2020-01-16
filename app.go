package app

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/supply"
	tmtypes "github.com/tendermint/tendermint/types"
	"os"
	"wings-blockchain/x/core"
	"wings-blockchain/x/currencies"
	"wings-blockchain/x/multisig"
	"wings-blockchain/x/vm"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"wings-blockchain/x/poa"
	poaTypes "wings-blockchain/x/poa/types"
)

const (
	appName = "wb"
)

type GenesisState map[string]json.RawMessage

var (
	// default home directories for the application CLI.
	DefaultCLIHome = os.ExpandEnv("$HOME/.wbcli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored.
	DefaultNodeHome = os.ExpandEnv("$HOME/.wbd")

	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{}, // genesis accounts management.
		genutil.AppModuleBasic{},     // utils to generate genesis transaction, genesis file validation, etc.
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distribution.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
		poa.AppModuleBasic{},
		currencies.AppModuleBasic{},
		multisig.AppModuleBasic{},
		vm.AppModuleBasic{},
	)

	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distribution.ModuleName:   nil,
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
	}
)

type WbServiceApp struct {
	*bam.BaseApp
	cdc      *codec.Codec
	msRouter core.Router

	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	accountKeeper  auth.AccountKeeper
	bankKeeper     bank.Keeper
	supplyKeeper   supply.Keeper
	paramsKeeper   params.Keeper
	stakingKeeper  staking.Keeper
	distrKeeper    distribution.Keeper
	slashingKeeper slashing.Keeper
	poaKeeper      poa.Keeper
	ccKeeper       currencies.Keeper
	msKeeper       multisig.Keeper
	vmKeeper       vm.Keeper

	mm *core.MsManager
}

// MakeCodec generates the necessary codecs for Amino.
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc) // register all module codecs.
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// NewWbServiceApp is a constructor function for wings blockchain.
func NewWbServiceApp(logger log.Logger, db dbm.DB, baseAppOptions ...func(*bam.BaseApp)) *WbServiceApp {
	cdc := MakeCodec()

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey,
		auth.StoreKey,
		supply.StoreKey,
		params.StoreKey,
		staking.StoreKey,
		distribution.StoreKey,
		slashing.StoreKey,
		poa.StoreKey,
		currencies.StoreKey,
		multisig.StoreKey,
		vm.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(
		params.TStoreKey,
		staking.TStoreKey,
	)

	var app = &WbServiceApp{
		BaseApp: bApp,
		cdc:     cdc,
		keys:    keys,
		tkeys:   tkeys,
	}

	// The ParamsKeeper handles parameter storage for the application.
	app.paramsKeeper = params.NewKeeper(app.cdc, keys[params.StoreKey], tkeys[params.TStoreKey], params.DefaultCodespace)

	// The AccountKeeper handles address -> account lookups.
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions.
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		app.ModuleAccountAddrs(),
	)

	// The SupplyKeeper collects transaction fees and renders them to the fee distribution module.
	app.supplyKeeper = supply.NewKeeper(cdc, keys[supply.StoreKey], app.accountKeeper, app.bankKeeper, maccPerms)

	// Initializing staking keeper.
	stakingKeeper := staking.NewKeeper(
		cdc,
		keys[staking.StoreKey],
		tkeys[staking.TStoreKey],
		app.supplyKeeper,
		app.paramsKeeper.Subspace(staking.DefaultParamspace),
		staking.DefaultCodespace,
	)

	// Initialize currency keeper.
	app.ccKeeper = currencies.NewKeeper(
		app.bankKeeper,
		keys[currencies.StoreKey],
		cdc,
	)

	// Initializing distribution keeper.
	app.distrKeeper = distribution.NewKeeper(
		cdc,
		keys[distribution.StoreKey],
		app.paramsKeeper.Subspace(distribution.DefaultParamspace),
		stakingKeeper,
		app.supplyKeeper,
		distribution.DefaultCodespace,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)

	// Initialize slashing keeper.
	app.slashingKeeper = slashing.NewKeeper(
		cdc,
		keys[slashing.StoreKey],
		stakingKeeper,
		app.paramsKeeper.Subspace(slashing.DefaultParamspace),
		slashing.DefaultCodespace,
	)

	// Initialize staking keeper.
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks(),
		),
	)

	// Initializing validators module.
	app.poaKeeper = poa.NewKeeper(
		keys[poa.StoreKey],
		app.cdc,
		app.paramsKeeper.Subspace(poaTypes.DefaultParamspace),
	)

	// Initializing multisignature router.
	app.msRouter = core.NewRouter()

	// Initializing multisignature router.
	app.msKeeper = multisig.NewKeeper(
		keys[multisig.StoreKey],
		app.cdc,
		app.msRouter,
	)

	// Initializing vm keeper.
	app.vmKeeper = vm.NewKeeper(
		keys[vm.StoreKey],
		app.cdc,
		app.paramsKeeper.Subspace(vm.DefaultParamspace),
	)

	// Initializing multisignature manager.
	app.mm = core.NewMsManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		distribution.NewAppModule(app.distrKeeper, app.supplyKeeper),
		staking.NewAppModule(app.stakingKeeper, app.distrKeeper, app.accountKeeper, app.supplyKeeper),
		poa.NewAppMsModule(app.poaKeeper),
		currencies.NewAppMsModule(app.ccKeeper),
		multisig.NewAppModule(app.msKeeper, app.poaKeeper),
		vm.NewAppModule(app.vmKeeper),
	)

	app.mm.SetOrderBeginBlockers(distribution.ModuleName, slashing.ModuleName)
	app.mm.SetOrderEndBlockers(staking.ModuleName, multisig.ModuleName)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		genaccounts.ModuleName,
		distribution.ModuleName,
		staking.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		slashing.ModuleName,
		supply.ModuleName,
		poa.ModuleName,
		currencies.ModuleName,
		multisig.ModuleName,
		vm.ModuleName,
		genutil.ModuleName,
	)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())
	app.mm.RegisterMsRoutes(app.msRouter)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.SetAnteHandler(
		auth.NewAnteHandler(
			app.accountKeeper,
			app.supplyKeeper,
			auth.DefaultSigVerificationGasConsumer,
		),
	)

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *WbServiceApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// Initialize chain function (initializing genesis data).
func (app *WbServiceApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState

	err := app.cdc.UnmarshalJSON(req.AppStateBytes, &genesisState)
	if err != nil {
		panic(err)
	}

	return app.mm.InitGenesis(ctx, genesisState)
}

// Initialize begin blocker function.
func (app *WbServiceApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// Initialize end blocker function.
func (app *WbServiceApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// Load app with specific height.
func (app *WbServiceApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// Exports genesis and validators.
func (app *WbServiceApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string,
) (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {

	// as if they could withdraw from the start of the next block.
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	genState := app.mm.ExportGenesis(ctx)
	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}

	validators = staking.WriteValidators(ctx, app.stakingKeeper)

	return appState, validators, nil
}
