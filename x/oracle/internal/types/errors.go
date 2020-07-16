package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal          = sdkErrors.Register(ModuleName, 0, "internal")
	ErrEmptyInput        = sdkErrors.Register(ModuleName, 1, "input must not be empty")
	ErrExpired           = sdkErrors.Register(ModuleName, 2, "price is expired")
	ErrNoValidPrice      = sdkErrors.Register(ModuleName, 3, "all input prices are expired")
	ErrInvalidAsset      = sdkErrors.Register(ModuleName, 4, "asset code not found")
	ErrInvalidOracle     = sdkErrors.Register(ModuleName, 5, "oracle not found or not authorized")
	ErrInvalidReceivedAt = sdkErrors.Register(ModuleName, 6, "invalid receivedAt")
	ErrExistingAsset     = sdkErrors.Register(ModuleName, 7, "asset code already exists")
)
