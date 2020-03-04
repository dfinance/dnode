package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeEmptyInput    sdk.CodeType = 1
	CodeExpired       sdk.CodeType = 2
	CodeInvalidPrice  sdk.CodeType = 3
	CodeInvalidAsset  sdk.CodeType = 4
	CodeInvalidOracle sdk.CodeType = 5
)

// Not used
func ErrEmptyInput(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyInput, fmt.Sprintf("input must not be empty"))
}

// New PostPrice is expired
func ErrExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpired, fmt.Sprintf("price is expired"))
}

// Not used
func ErrNoValidPrice(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPrice, fmt.Sprintf("all input prices are expired"))
}

// Asset not found
func ErrInvalidAsset(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAsset, fmt.Sprintf("asset code not found"))
}

// Asset already exists
func ErrExistingAsset(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAsset, fmt.Sprintf("asset code already exists"))
}

// Oracle not found
func ErrInvalidOracle(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidOracle, fmt.Sprintf("oracle not found or not authorized"))
}
