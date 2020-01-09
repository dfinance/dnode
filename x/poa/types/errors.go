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

// When validator already exists
func ErrValidatorExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorExists, "validator already exists %s", address)
}

// When validator doesnt exists
func ErrValidatorDoesntExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorDoesntExist, "validator doesn't exist %s", address)
}

// When validators maximum limit reached
func ErrMaxValidatorsReached(max uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMaxValidatorsReached, "maxium %d validators reached", max)
}

// When validators minimum limit reached
func ErrMinValidatorsReached(min uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMinValidatorsReached, "minimum %d validators reached", min)
}

// When validator's ethereum address is wrong
func ErrWrongEthereumAddress(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeWrongEthereumAddress, "wrong ethereum address %s for validator", address)
}

// When not enough validators to initialize genesis
func ErrNotEnoungValidators(actual uint16, min uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeNotEnoughValidators, "%d not enough validators to init genesis, min is %d", actual, min)
}
