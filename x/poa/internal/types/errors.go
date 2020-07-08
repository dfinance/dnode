package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal             = sdkErrors.Register(ModuleName, 100, "internal")
	ErrWrongEthereumAddress = sdkErrors.Register(ModuleName, 101, "wrong ethereum address for validator")
	ErrValidatorExists      = sdkErrors.Register(ModuleName, 102, "validator already exists")
	ErrValidatorNotExists   = sdkErrors.Register(ModuleName, 103, "validator not found")
	ErrMaxValidatorsReached = sdkErrors.Register(ModuleName, 104, "maximum number of validators reached")
	ErrMinValidatorsReached = sdkErrors.Register(ModuleName, 105, "minimum number of validators reached")
)
