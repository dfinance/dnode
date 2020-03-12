// Describing errors and codes.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeValidatorExists      sdk.CodeType = 101
	CodeValidatorDoesntExist sdk.CodeType = 102
	CodeMaxValidatorsReached sdk.CodeType = 103
	CodeMinValidatorsReached sdk.CodeType = 104

	CodeWrongEthereumAddress sdk.CodeType = 201

	CodeNotEnoughValidators sdk.CodeType = 301
)

// Msg.Validator already exists
func ErrValidatorExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorExists, "%q validator already exists", address)
}

// Msg.Validator not found
func ErrValidatorDoesntExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorDoesntExist, "%q validator not found", address)
}

// Validators maximum limit reached (on genesis init / add validator request)
func ErrMaxValidatorsReached(max uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMaxValidatorsReached, "maximum %d validators reached", max)
}

// Validators minimum limit reached (on genesis init / add validator request)
func ErrMinValidatorsReached(min uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMinValidatorsReached, "minimum %d validators reached", min)
}

// Validator's ethereum address is invalid (on validator add / replace)
func ErrWrongEthereumAddress(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeWrongEthereumAddress, "wrong ethereum address %q for validator", address)
}

// Not enough validators to initialize genesis
func ErrNotEnoungValidators(actual uint16, min uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotEnoughValidators, "%d validators is not enough to init genesis, min is %d", actual, min)
}
