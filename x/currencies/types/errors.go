package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)
const (
	CodeErrWrongSymbol   = 101
	CodeErrWrongAmount   = 102
	CodeErrWrongDecimals = 103
)

func ErrWrongSymbol(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongSymbol, "wrong symbol %s", symbol)
}

func ErrWrongAmount(amount int64) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongAmount, "wrong amount %d, should be " +
		"great then zero", amount)
}

func ErrWrongDecimals(decimals int8) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongDecimals, "%d decimals or amount can't be less/equal 0, " +
		"and decimals should be less then 8", decimals)
}