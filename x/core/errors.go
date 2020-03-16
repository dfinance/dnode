package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Codespace sdk.CodespaceType = "core"

	CodeFeeRequired   sdk.CodeType = 101
	CodeWrongFeeDenom sdk.CodeType = 102
)

// StdTx Fee.Amount is empty
func ErrFeeRequired() sdk.Error {
	return sdk.NewError(Codespace, CodeFeeRequired, "tx must contain fees")
}

// StdTx Fee.Amount wrong denom
func ErrWrongFeeDenom(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeWrongFeeDenom, "tx must contain fees only in %q denom", denom)
}
