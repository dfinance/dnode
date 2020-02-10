package oracle

import (
	"wings-blockchain/x/oracle/internal/keeper"
	"wings-blockchain/x/oracle/internal/types"
)

type (
	GenesisState       = types.GenesisState
	MsgPostPrice       = types.MsgPostPrice
	Params             = types.Params
	ParamSubspace      = types.ParamSubspace
	QueryRawPricesResp = types.QueryRawPricesResp
	QueryAssetsResp    = types.QueryAssetsResp
	Asset              = types.Asset
	Assets             = types.Assets
	Oracle             = types.Oracle
	Oracles            = types.Oracles
	CurrentPrice       = types.CurrentPrice
	PostedPrice        = types.PostedPrice
	SortDecs           = types.SortDecs
	Keeper             = keeper.Keeper
	MsgAddOracle       = types.MsgAddOracle
)

const (
	DefaultCodespace  = types.DefaultCodespace
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultParamspace
	StoreKey          = types.StoreKey
)

var (
	ModuleCdc     = types.ModuleCdc
	NewKeeper     = keeper.NewKeeper
	RegisterCodec = types.RegisterCodec
	// functions aliases
	ErrEmptyInput       = types.ErrEmptyInput
	ErrExpired          = types.ErrExpired
	ErrNoValidPrice     = types.ErrNoValidPrice
	ErrInvalidAsset     = types.ErrInvalidAsset
	ErrInvalidOracle    = types.ErrInvalidOracle
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	NewMsgPostPrice     = types.NewMsgPostPrice
	ParamKeyTable       = types.ParamKeyTable
	NewParams           = types.NewParams
	DefaultParams       = types.DefaultParams
	NewQuerier          = keeper.NewQuerier
)
