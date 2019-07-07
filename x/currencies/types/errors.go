package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeErrWrongSymbol   	 = 101
	CodeErrWrongAmount   	 = 102
	CodeErrWrongDecimals 	 = 103
	CodeErrWrongIssueID      = 104
	CodeErrIncorrectDecimals = 105
	CodeErrExistsIssue       = 106
	CodeErrNotExistCurrency  = 107
)

func ErrWrongSymbol(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongSymbol, "wrong symbol %s", symbol)
}

func ErrWrongAmount(amount string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongAmount, "wrong amount %s, should be " +
		"great then zero", amount)
}

func ErrWrongDecimals(decimals int8) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongDecimals, "%d decimals can't be less/equal 0 ", decimals)
}

func ErrWrongIssueID(issueID string) sdk.Error {
    return sdk.NewError(DefaultCodespace, CodeErrWrongIssueID, "%s is wrong issue id", issueID)
}

func ErrIncorrectDecimals(d1, d2 int8, symbol string) sdk.Error {
    return sdk.NewError(DefaultCodespace, CodeErrIncorrectDecimals, "currency %s must have %d " +
        "decimals instead of %d decimals", symbol, d1, d2)
}

func ErrExistsIssue(issueID string) sdk.Error {
    return sdk.NewError(DefaultCodespace, CodeErrExistsIssue, "issue with %s id already exists", issueID)
}

func ErrNotExistCurrency(symbol string) sdk.Error {
    return sdk.NewError(DefaultCodespace, CodeErrNotExistCurrency, "currency %s doesnt exist yet", symbol)
}
