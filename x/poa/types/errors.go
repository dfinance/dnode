// Describing errors and codes.
package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal = sdkErrors.Register(ModuleName, 100, "internal")
	// Msg.Validator already exists.
	ErrValidatorExists = sdkErrors.Register(ModuleName, 101, "validator already exists")
	// Msg.Validator not found.
	ErrValidatorDoesntExists = sdkErrors.Register(ModuleName, 102, "validator not found")
	// Validators maximum limit reached (on genesis init / add validator request).
	ErrMaxValidatorsReached = sdkErrors.Register(ModuleName, 103, "maximum number of validators reached")
	// Validators minimum limit reached (on genesis init / add validator request).
	ErrMinValidatorsReached = sdkErrors.Register(ModuleName, 104, "minimum number of validators reached")

	// Validator's ethereum address is invalid (on validator add / replace).
	ErrWrongEthereumAddress = sdkErrors.Register(ModuleName, 201, "wrong ethereum address for validator")

	// Not enough validators to initialize genesis.
	ErrNotEnoungValidators = sdkErrors.Register(ModuleName, 301, "number of validators is not enough to init genesis")
)
