package oracle

import (
	"github.com/dfinance/dnode/x/oracle/internal/keeper"
	"github.com/dfinance/dnode/x/oracle/internal/types"
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
	MsgSetOracles      = types.MsgSetOracles
	MsgAddAsset        = types.MsgAddAsset
	MsgSetAsset        = types.MsgSetAsset
	PostPriceParams    = types.PostPriceParams
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultParamspace
	StoreKey          = types.StoreKey
)

var (
	ModuleCdc     = types.ModuleCdc
	NewKeeper     = keeper.NewKeeper
	NewAsset      = types.NewAsset
	RegisterCodec = types.RegisterCodec
	// functions aliases
	ErrEmptyInput       = types.ErrEmptyInput
	ErrExpired          = types.ErrExpired
	ErrNoValidPrice     = types.ErrNoValidPrice
	ErrInvalidAsset     = types.ErrInvalidAsset
	ErrInvalidOracle    = types.ErrInvalidOracle
	DefaultGenesisState = types.DefaultGenesisState
	NewMsgPostPrice     = types.NewMsgPostPrice
	ParamKeyTable       = types.ParamKeyTable
	NewParams           = types.NewParams
	DefaultParams       = types.DefaultParams
	NewQuerier          = keeper.NewQuerier
)
