package vmauth

import (
	"github.com/cosmos/cosmos-sdk/x/auth"
	authClientCli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authClientRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dfinance/dnode/x/vmauth/internal/keeper"
	"github.com/dfinance/dnode/x/vmauth/internal/types"
)

type (
	Keeper       = keeper.VMAccountKeeper
	GenesisState = authTypes.GenesisState
)

const (
	ModuleName   = types.ModuleName
	QuerierRoute = authTypes.QuerierRoute
)

var (
	// variable aliases
	ModuleCdc = authTypes.ModuleCdc
	// function aliases
	RegisterCodec       = authTypes.RegisterCodec
	NewKeeper           = keeper.NewKeeper
	ValidateGenesis     = authTypes.ValidateGenesis
	ExportGenesis       = auth.ExportGenesis
	RegisterRoutes      = authClientRest.RegisterRoutes
	GetTxCmd            = authClientCli.GetTxCmd
	GetQueryCmd         = authClientCli.GetQueryCmd
	DefaultGenesisState = authTypes.DefaultGenesisState
	// perms requests
	RequestCCStoragePerms = types.RequestCCStoragePerms
)
