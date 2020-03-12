// Implements errors codes and functions for currencies module.
package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeErrWrongSymbol       = 101
	CodeErrWrongAmount       = 102
	CodeErrWrongDecimals     = 103
	CodeErrWrongIssueID      = 104
	CodeErrIncorrectDecimals = 105
	CodeErrExistsIssue       = 106
	CodeErrNotExistCurrency  = 107
	CodeErrWrongRecipient    = 108
)

// Msg.Symbol is empty
func ErrWrongSymbol(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongSymbol, "wrong symbol %q", symbol)
}

// Msg.Amount is zero
func ErrWrongAmount(amount string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongAmount, "wrong amount %q, should be "+
		"greater than zero", amount)
}

// Msg.Decimals < 0
func ErrWrongDecimals(decimals int8) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongDecimals, "%d decimals can't be less than 0 ", decimals)
}

// Issue.Recipient is empty / Msg.IssueID is empty
func ErrWrongIssueID(issueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongIssueID, "wrong issueID %q", issueID)
}

// Currency.Decimals != decimals in request
func ErrIncorrectDecimals(d1, d2 int8, symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrIncorrectDecimals, "currency %q must have %d "+
		"decimals instead of %d decimals", symbol, d1, d2)
}

// IssueID already exists in store
func ErrExistsIssue(issueID string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrExistsIssue, "issueID %q already exists", issueID)
}

// Currency.Symbol != requested symbol
func ErrNotExistCurrency(symbol string) sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrNotExistCurrency, "currency %q not found", symbol)
}

// Msg.Recipient is empty
func ErrWrongRecipient() sdk.Error {
	return sdk.NewError(DefaultCodespace, CodeErrWrongRecipient, "empty recipient is not allowed")
}