package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	DefaultCodespace sdk.CodespaceType = ModuleName

	CodeEmptyInput        sdk.CodeType = 1
	CodeExpired           sdk.CodeType = 2
	CodeInvalidPrice      sdk.CodeType = 3
	CodeInvalidAsset      sdk.CodeType = 4
	CodeInvalidOracle     sdk.CodeType = 5
	CodeInvalidReceivedAt sdk.CodeType = 6
)

// Not used
func ErrEmptyInput(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyInput, "input must not be empty")
}

// New PostPrice is expired
func ErrExpired(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpired, "price is expired")
}

// New PostPrice ReceivedAt field is invalid
func ErrInvalidReceivedAt(codespace sdk.CodespaceType, comment string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidReceivedAt, fmt.Sprintf("Invalid receivedAt: %s.", comment))
}

// Not used
func ErrNoValidPrice(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPrice, "all input prices are expired")
}

// Asset not found
func ErrInvalidAsset(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAsset, "asset code not found")
}

// Asset already exists
func ErrExistingAsset(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidAsset, "asset code already exists")
}

// Oracle not found
func ErrInvalidOracle(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidOracle, "oracle not found or not authorized")
}
