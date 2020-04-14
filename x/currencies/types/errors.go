// Implements errors codes and functions for currencies module.
package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal          = sdkErrors.Register(ModuleName, 100, "internal")
	// Msg.Symbol is empty.
	ErrWrongSymbol       = sdkErrors.Register(ModuleName, 101, "wrong symbol")
	// Msg.Amount is zero.
	ErrWrongAmount       = sdkErrors.Register(ModuleName, 102, "wrong amount, should be greater that 0")
	// Msg.Decimals < 0.
	ErrWrongDecimals     = sdkErrors.Register(ModuleName, 103, "decimals can't be less than 0")
	// Issue.Recipient is empty / Msg.IssueID is empty.
	ErrWrongIssueID      = sdkErrors.Register(ModuleName, 104, "wrong issueID")
	// Currency.Decimals != decimals in request.
	ErrIncorrectDecimals = sdkErrors.Register(ModuleName, 105, "currency decimals should match")
	// IssueID already exists in store.
	ErrExistsIssue       = sdkErrors.Register(ModuleName, 106, "issueID already exists")
	// Currency.Symbol != requested symbol.
	ErrNotExistCurrency  = sdkErrors.Register(ModuleName, 107, "currency not found")
	// Msg.Recipient is empty.
	ErrWrongRecipient    = sdkErrors.Register(ModuleName, 108, "empty recipient is not allowed")
)
