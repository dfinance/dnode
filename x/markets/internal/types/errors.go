package types

import sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// ID is invalid or not found.
	ErrWrongID = sdkErrors.Register(ModuleName, 101, "wrong ID")
	// AssetDenom is empty.
	ErrWrongAssetDenom = sdkErrors.Register(ModuleName, 102, "wrong asset denom")
	// Market already exists.
	ErrMarketExists = sdkErrors.Register(ModuleName, 103, "market exists")
	// Base to Quote asset quantity convert failed.
	ErrInvalidQuantity = sdkErrors.Register(ModuleName, 104, "base to quote asset quantity normalization failed")
	// MsgCreateMarket.From is empty.
	ErrWrongFrom = sdkErrors.Register(ModuleName, 105, "wrong from address, should not be empty")
)
