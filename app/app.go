package app

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmOs "github.com/tendermint/tendermint/libs/os"
	tmTypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/genaccounts"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orderbook"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vmauth"
)

const (
	appName = "dfinance" // application name.
)

type GenesisState map[string]json.RawMessage

var (
	// default home directories for the application CLI.
	DefaultCLIHome = os.ExpandEnv("$HOME/.dncli")

	// DefaultNodeHome sets the folder where the applcation data and configuration will be stored.
	DefaultNodeHome = os.ExpandEnv("$HOME/.dnode")

	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distribution.AppModuleBasic{},
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		evidence.AppModuleBasic{},
		supply.AppModuleBasic{},
		poa.AppModuleBasic{},
		ccstorage.AppModuleBasic{},
		currencies.AppModuleBasic{},
		multisig.AppModuleBasic{},
		oracle.AppModuleBasic{},
		gov.AppModuleBasic{},
		vm.AppModuleBasic{},
		markets.AppModuleBasic{},
		orders.AppModuleBasic{},
		orderbook.AppModuleBasic{},
	)

	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distribution.ModuleName:   nil,
		mint.ModuleName:           {supply.Minter},
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		gov.ModuleName:            {supply.Burner},
		orders.ModuleName:         {supply.Burner},
	}
)

// DN Service App implements DN mains logic.
type DnServiceApp struct {
	*BaseApp

	cdc       *codec.Codec
	msRouter  msmodule.MsRouter
	govRouter govTypes.Router

	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	accountKeeper   vmauth.Keeper
	bankKeeper      bank.Keeper
	supplyKeeper    supply.Keeper
	paramsKeeper    params.Keeper
	stakingKeeper   staking.Keeper
	mintKeeper      mint.Keeper
	distrKeeper     distribution.Keeper
	slashingKeeper  slashing.Keeper
	evidenceKeeper  evidence.Keeper
	poaKeeper       poa.Keeper
	ccsKeeper       ccstorage.Keeper
	ccKeeper        currencies.Keeper
	msKeeper        multisig.Keeper
	vmKeeper        vm.Keeper
	oracleKeeper    oracle.Keeper
	govKeeper       gov.Keeper
	marketKeeper    markets.Keeper
	orderKeeper     orders.Keeper
	orderBookKeeper orderbook.Keeper

	mm *msmodule.MsManager

	// vm connection
	vmConn     *grpc.ClientConn
	vmListener net.Listener
}

// Initialize connection to VM server.
func (app *DnServiceApp) InitializeVMConnection(addr string) {
	var err error

	app.Logger().Info(fmt.Sprintf("Creating connection to VM, address: %s", addr))
	app.vmConn, err = helpers.GetGRpcClientConnection(addr, 1*time.Second)
	if err != nil {
		panic(err)
	}

	app.Logger().Info(fmt.Sprintf("Non-blocking connection initialized, status: %s", app.vmConn.GetState()))
}

// Close VM connection and DS server stops.
func (app DnServiceApp) CloseConnections() {
	app.vmKeeper.CloseConnections()
}

// Initialize listener to listen for connections from VM for data server.
func (app *DnServiceApp) InitializeVMDataServer(addr string) {
	var err error

	app.Logger().Info(fmt.Sprintf("Starting VM data server listener, address: %s", addr))
	app.vmListener, err = helpers.GetGRpcNetListener(addr)
	if err != nil {
		panic(err)
	}

	app.Logger().Info("VM data server is running")
}

// MakeCodec generates the necessary codecs for Amino.
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc) // register all module codecs.
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)
	return cdc
}

// NewDnServiceApp is a constructor function for dfinance blockchain.
func NewDnServiceApp(logger log.Logger, db dbm.DB, config *config.VMConfig, baseAppOptions ...func(*BaseApp)) *DnServiceApp {
	cdc := MakeCodec()

	bApp := NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetAppVersion(version.Version)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey,
		auth.StoreKey,
		supply.StoreKey,
		params.StoreKey,
		staking.StoreKey,
		mint.StoreKey,
		distribution.StoreKey,
		slashing.StoreKey,
		evidence.StoreKey,
		poa.StoreKey,
		ccstorage.StoreKey,
		currencies.StoreKey,
		multisig.StoreKey,
		vm.StoreKey,
		oracle.StoreKey,
		gov.StoreKey,
		orders.StoreKey,
		orderbook.StoreKey,
	)

	tkeys := sdk.NewTransientStoreKeys(
		params.TStoreKey,
		staking.TStoreKey,
	)

	var app = &DnServiceApp{
		BaseApp: bApp,
		cdc:     cdc,
		keys:    keys,
		tkeys:   tkeys,
	}

	// initialize connections
	app.InitializeVMDataServer(config.DataListen)
	app.InitializeVMConnection(config.Address)

	// Reduce ConsensusPower reduction coefficient (1 dfi == 1 power unit)
	// 1 dfi == 1000000000000000000
	sdk.PowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))

	// The ParamsKeeper handles parameter storage for the application.
	app.paramsKeeper = params.NewKeeper(
		app.cdc,
		keys[params.StoreKey],
		tkeys[params.TStoreKey],
	)

	// Initializing vm keeper.
	var err error
	app.vmKeeper = vm.NewKeeper(
		keys[vm.StoreKey],
		cdc,
		app.vmConn,
		app.vmListener,
		config,
	)

	// Initializing currencies storage keeper.
	app.ccsKeeper = ccstorage.NewKeeper(
		cdc,
		keys[ccstorage.StoreKey],
		app.paramsKeeper.Subspace(ccstorage.DefaultParamspace),
		app.vmKeeper,
	)

	// The AccountKeeper handles address -> account lookups.
	app.accountKeeper = vmauth.NewKeeper(
		cdc,
		keys[auth.StoreKey],
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		app.ccsKeeper,
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions.
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		app.ModuleAccountAddrs(),
	)

	// The SupplyKeeper collects transaction fees and renders them to the fee distribution module.
	app.supplyKeeper = supply.NewKeeper(
		cdc,
		keys[supply.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		maccPerms,
	)

	// Initializing staking keeper.
	stakingKeeper := staking.NewKeeper(
		cdc,
		keys[staking.StoreKey],
		app.supplyKeeper,
		app.paramsKeeper.Subspace(staking.DefaultParamspace),
	)

	// Mint keeper (inflation).
	app.mintKeeper = mint.NewKeeper(
		cdc, keys[mint.StoreKey],
		app.paramsKeeper.Subspace(mint.DefaultParamspace),
		stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName,
	)

	// Evidence keeper. Catch double sign and provide evidence to confirm Byzantine validators.
	evidenceKeeper := evidence.NewKeeper(
		cdc,
		keys[evidence.StoreKey],
		app.paramsKeeper.Subspace(evidence.DefaultParamspace),
		stakingKeeper,
		app.slashingKeeper,
	)
	evidenceRouter := evidence.NewRouter()
	evidenceKeeper.SetRouter(evidenceRouter)
	app.evidenceKeeper = *evidenceKeeper

	// Initialize currency keeper.
	app.ccKeeper = currencies.NewKeeper(
		cdc,
		keys[currencies.StoreKey],
		app.bankKeeper,
		app.supplyKeeper,
		app.ccsKeeper,
	)

	// Initializing distribution keeper.
	app.distrKeeper = distribution.NewKeeper(
		cdc,
		keys[distribution.StoreKey],
		app.paramsKeeper.Subspace(distribution.DefaultParamspace),
		stakingKeeper,
		app.supplyKeeper,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)

	// Initialize slashing keeper.
	app.slashingKeeper = slashing.NewKeeper(
		cdc,
		keys[slashing.StoreKey],
		stakingKeeper,
		app.paramsKeeper.Subspace(slashing.DefaultParamspace),
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
		cdc,
		keys[poa.StoreKey],
		app.paramsKeeper.Subspace(poa.DefaultParamspace),
	)

	// Initializing multisignature router.
	app.msRouter = msmodule.NewMsRouter()

	// Initializing multisignature router.
	app.msKeeper = multisig.NewKeeper(
		cdc,
		keys[multisig.StoreKey],
		app.paramsKeeper.Subspace(multisig.DefaultParamspace),
		app.msRouter,
		app.poaKeeper,
	)

	// Initializing oracle module.
	app.oracleKeeper = oracle.NewKeeper(
		keys[oracle.StoreKey],
		cdc,
		app.paramsKeeper.Subspace(oracle.DefaultParamspace),
		app.vmKeeper,
	)

	// The Governance keeper.
	app.govRouter = gov.NewRouter()
	app.govRouter.AddRoute(vm.GovRouterKey, vm.NewGovHandler(app.vmKeeper))
	app.govRouter.AddRoute(currencies.GovRouterKey, currencies.NewGovHandler(app.ccKeeper))

	app.govKeeper = gov.NewKeeper(
		cdc,
		keys[gov.StoreKey],
		app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable()),
		app.supplyKeeper,
		app.stakingKeeper,
		app.govRouter,
	)

	// Initializing markets module.
	app.marketKeeper = markets.NewKeeper(
		cdc,
		app.paramsKeeper.Subspace(markets.DefaultParamspace),
		app.ccsKeeper,
	)

	// Initializing orders module.
	app.orderKeeper = orders.NewKeeper(
		keys[orders.StoreKey],
		cdc,
		app.bankKeeper,
		app.supplyKeeper,
		app.marketKeeper,
	)

	// Initializing order_bool module.
	app.orderBookKeeper = orderbook.NewKeeper(
		keys[orderbook.StoreKey],
		cdc,
		app.orderKeeper,
	)

	// Initializing multi signature manager.
	app.mm = msmodule.NewMsManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		vmauth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.stakingKeeper),
		distribution.NewAppModule(app.distrKeeper, app.accountKeeper, app.supplyKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.supplyKeeper),
		mint.NewAppModule(app.mintKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		poa.NewAppMsModule(app.poaKeeper),
		ccstorage.NewAppModule(app.ccsKeeper),
		currencies.NewAppMsModule(app.ccKeeper, app.ccsKeeper),
		multisig.NewAppModule(app.msKeeper),
		oracle.NewAppModule(app.oracleKeeper),
		vm.NewAppModule(app.vmKeeper),
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.supplyKeeper),
		markets.NewAppModule(app.marketKeeper),
		orders.NewAppModule(app.orderKeeper),
		orderbook.NewAppModule(app.orderBookKeeper),
	)

	app.mm.SetOrderBeginBlockers(
		mint.ModuleName,
		currencies.ModuleName, // Must go after mint.
		distribution.ModuleName,
		slashing.ModuleName,
		vm.ModuleName,
	)
	app.mm.SetOrderEndBlockers(
		gov.ModuleName,
		staking.ModuleName,
		multisig.ModuleName,
		oracle.ModuleName,
		orders.ModuleName,
		orderbook.ModuleName,
	)

	// Sets the order of Genesis - Order matters, genutil is to always come last
	// NOTE: The genutils moodule must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		vm.ModuleName,
		ccstorage.ModuleName,
		genaccounts.ModuleName,
		distribution.ModuleName,
		staking.ModuleName,
		auth.ModuleName,
		bank.ModuleName,
		slashing.ModuleName,
		gov.ModuleName,
		poa.ModuleName,
		multisig.ModuleName,
		currencies.ModuleName,
		oracle.ModuleName,
		markets.ModuleName,
		orders.ModuleName,
		orderbook.ModuleName,
		mint.ModuleName,
		evidence.ModuleName,
		supply.ModuleName, // should be after all modules related to account balances.
		genutil.ModuleName,
	)

	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())
	app.mm.RegisterMsRoutes(app.msRouter)

	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)

	app.SetAnteHandler(
		core.NewAnteHandler(
			app.accountKeeper,
			app.supplyKeeper,
			auth.DefaultSigVerificationGasConsumer,
		),
	)

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	err = app.LoadLatestVersion(app.keys[bam.MainStoreKey])
	if err != nil {
		tmOs.Exit(err.Error())
	}

	// Temporary solution, but seems works.
	// Set context for reading data from DS store.
	// TODO: find another way for storage to read data.
	dsContext := app.GetDSContext()
	app.vmKeeper.SetDSContext(dsContext)
	app.vmKeeper.StartDSServer(dsContext)
	time.Sleep(1 * time.Second) // need for DS to initialize stdlib, will be removed later.

	return app
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *DnServiceApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// Initialize chain function (initializing genesis data).
func (app *DnServiceApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState

	err := app.cdc.UnmarshalJSON(req.AppStateBytes, &genesisState)
	if err != nil {
		panic(err)
	}

	resp := app.mm.InitGenesis(ctx, genesisState)
	app.vmKeeper.SetDSContext(ctx)
	app.vmKeeper.StartDSServer(ctx)
	time.Sleep(1 * time.Second) // need for DS to initialize stdlib, will be removed later.

	return resp
}

// Initialize begin blocker function.
func (app *DnServiceApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// Initialize end blocker function.
func (app *DnServiceApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// Load app with specific height.
func (app *DnServiceApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// Exports genesis and validators.
func (app *DnServiceApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string,
) (appState json.RawMessage, validators []tmTypes.GenesisValidator, err error) {

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
