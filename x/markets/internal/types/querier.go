package types

import (
	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryList   = "list"
	QueryMarket = "market"
)

// Client request for market.
type MarketReq struct {
	ID dnTypes.ID `json:"id" yaml:"id"`
}

// Client request for markets.
type MarketsReq struct {
	// Page number
	Page int
	// Items per page
	Limit int
	// BaseAsset denom filter
	BaseAssetDenom string
	// QuoteAsset denom filter
	QuoteAssetDenom string
	// AssetCode filter
	AssetCode string
}

// NewMarketsFilter returned MarketsReq object with filled required fields page and limit.
func NewMarketsFilter(page, limit int) MarketsReq {
	return MarketsReq{
		Page:  page,
		Limit: limit,
	}
}

// BaseDenomFilter check if BaseAssetDenom filter is enabled.
func (r MarketsReq) BaseDenomFilter() bool {
	return r.BaseAssetDenom != ""
}

// QuoteDenomFilter check if QuoteAssetDenom filter is enabled.
func (r MarketsReq) QuoteDenomFilter() bool {
	return r.QuoteAssetDenom != ""
}

// AssetCodeFilter check if AssetCode filter is enabled.
func (r MarketsReq) AssetCodeFilter() bool {
	return r.AssetCode != ""
}
