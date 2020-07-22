package oracle

import (
	"github.com/dfinance/dnode/x/oracle/internal/keeper"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

type (
	GenesisState       = types.GenesisState
	MsgPostPrice       = types.MsgPostPrice
	Params             = types.Params
	QueryRawPricesResp = types.QueryRawPricesResp
	QueryAssetsResp    = types.QueryAssetsResp
	Asset              = types.Asset
	Assets             = types.Assets
	Oracle             = types.Oracle
	Oracles            = types.Oracles
	CurrentPrice       = types.CurrentPrice
	PostedPrice        = types.PostedPrice
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
	DefaultParamspace = types.DefaultParamspace
	StoreKey          = types.StoreKey
	//
	QueryAssets    = types.QueryAssets
	QueryRawPrices = types.QueryRawPrices
	QueryPrice     = types.QueryPrice
	// Event types, attribute types and values
	EventTypePrice = types.EventTypePrice
	//
	AttributeAssetCode  = types.AttributeAssetCode
	AttributePrice      = types.AttributePrice
	AttributeReceivedAt = types.AttributeReceivedAt
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// functions aliases
	RegisterCodec       = types.RegisterCodec
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	DefaultGenesisState = types.DefaultGenesisState
	DefaultParams       = types.DefaultParams
	NewParams           = types.NewParams
	NewAsset            = types.NewAsset
	NewMsgPostPrice     = types.NewMsgPostPrice
	// perms requests
	RequestVMStoragePerms = types.RequestVMStoragePerms
	// errors
	ErrInternal      = types.ErrInternal
	ErrEmptyInput    = types.ErrEmptyInput
	ErrExpired       = types.ErrExpired
	ErrNoValidPrice  = types.ErrNoValidPrice
	ErrExistingAsset = types.ErrExistingAsset
	ErrInvalidAsset  = types.ErrInvalidAsset
	ErrInvalidOracle = types.ErrInvalidOracle
)
