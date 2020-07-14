package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	EventTypeClearance        = ModuleName + ".clearance"
	EventAttributeKeyMarketID = "market_id"
	EventAttributeKeyPrice    = "price"
)

func NewClearanceEvent(marketID dnTypes.ID, clearancePrice sdk.Uint) sdk.Event {
	return sdk.NewEvent(
		EventTypeClearance,
		sdk.NewAttribute(EventAttributeKeyMarketID, marketID.String()),
		sdk.NewAttribute(EventAttributeKeyPrice, clearancePrice.String()),
	)
}
