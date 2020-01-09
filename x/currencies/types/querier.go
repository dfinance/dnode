// Querier.
package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Request to print destroys
type DestroysReq struct {
	Page  sdk.Int
	Limit sdk.Int
}

// Request to get destroy by destroy id.
type DestroyReq struct {
	DestroyId sdk.Int
}

// Request to get issue by id.
type IssueReq struct {
	IssueID string
}

// Request to get currency by id.
type CurrencyReq struct {
	Symbol string
}
