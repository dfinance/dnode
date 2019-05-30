package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeValidatorExists 		sdk.CodeType = 101
	CodeValidatorDoesntExist 	sdk.CodeType = 102

	CodeMaxValidatorsReached	sdk.CodeType = 202
	CodeMinValidatorsReached 	sdk.CodeType = 203
)

func ErrValidatorExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorExists, "validator already exists %s", address)
}

func ErrValidatorDoesntExists(address string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeValidatorDoesntExist, "validator doesn't exist %s", address)
}

func ErrMaxValidatorsReached(max uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMaxValidatorsReached, "maxium %d validators reached", max)
}

func ErrMinValidatorsReached(min uint16) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeMinValidatorsReached, "minimum %d validators reached", min)
}