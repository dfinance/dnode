package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal = sdkErrors.Register(ModuleName, 0, "internal")
	// Not used.
	ErrEmptyInput = sdkErrors.Register(ModuleName, 1, "input must not be empty")
	// New PostPrice is expired.
	ErrExpired = sdkErrors.Register(ModuleName, 2, "price is expired")
	// Not used.
	ErrNoValidPrice = sdkErrors.Register(ModuleName, 3, "all input prices are expired")
	// Asset not found.
	ErrInvalidAsset = sdkErrors.Register(ModuleName, 4, "asset code not found")
	// Oracle not found.
	ErrInvalidOracle = sdkErrors.Register(ModuleName, 5, "oracle not found or not authorized")
	// New PostPrice ReceivedAt field is invalid.
	ErrInvalidReceivedAt = sdkErrors.Register(ModuleName, 6, "invalid receivedAt")
	// Asset already exists.
	ErrExistingAsset = sdkErrors.Register(ModuleName, 7, "asset code already exists")
)
