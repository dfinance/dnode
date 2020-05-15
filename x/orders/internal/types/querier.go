package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client request for order.
type OrderReq struct {
	ID dnTypes.ID `json:"id" yaml:"id"`
}

// Client request for markets.
type OrdersReq struct {
	// Page number
	Page  int
	// Items per page
	Limit int
	// Owner filter
	Owner sdk.AccAddress
	// Direction filter
	Direction Direction
	// MarketID filter
	MarketID string
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
