package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	Codespace sdk.CodespaceType = "core"

	CodeFeeRequired   sdk.CodeType = 101 // When fee is zero
	CodeWrongFeeDenom sdk.CodeType = 102 // When fee denom is wrong
)

func ErrFeeRequired() sdk.Error {
	return sdk.NewError(Codespace, CodeFeeRequired, "tx must contains fees")
}

func ErrWrongFeeDenom(denom string) sdk.Error {
	return sdk.NewError(Codespace, CodeWrongFeeDenom, "tx must contains fees only in %s denom", denom)
}
