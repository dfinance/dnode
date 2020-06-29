package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryDestroys = "destroys"
	QueryDestroy  = "destroy"
	QueryIssue    = "issue"
	QueryCurrency = "currency"
)

// Client request for destroy.
type DestroyReq struct {
	ID dnTypes.ID
}

// Client request for destroys.
type DestroysReq struct {
	Page  sdk.Uint
	Limit sdk.Uint
}

// Client request for issue.
type IssueReq struct {
	ID string
}

// Client request for currency.
type CurrencyReq struct {
	Denom string
}
