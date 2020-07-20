package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryList  = "list"
	QueryOrder = "order"
)

// Client request for order.
type OrderReq struct {
	ID dnTypes.ID `json:"id" yaml:"id"`
}

// Client request for markets.
type OrdersReq struct {
	// Page number
	Page  sdk.Uint `json:"page" yaml:"page"`
	// Items per page
	Limit sdk.Uint `json:"limit" yaml:"limit"`
	// Owner filter
	Owner sdk.AccAddress `json:"owner" yaml:"owner"`
	// Direction filter
	Direction Direction `json:"direction" yaml:"direction"`
	// MarketID filter
	MarketID string `json:"market_id" yaml:"market_id"`
}

// OwnerFilter check if Owner filter is enabled.
func (r OrdersReq) OwnerFilter() bool {
	return !r.Owner.Empty()
}

// OwnerFilter check if Direction filter is enabled.
func (r OrdersReq) DirectionFilter() bool {
	return r.Direction.IsValid()
}

// OwnerFilter check if MarketID filter is enabled.
func (r OrdersReq) MarketIDFilter() bool {
	return r.MarketID != ""
}
