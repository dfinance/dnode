package types

import (
	"strings"
)

// price Takes an [assetcode] and returns CurrentPrice for that asset
// oracle Takes an [assetcode] and returns the raw []PostedPrice for that asset
// assets Returns []Assets in the oracle system

const (
	QueryPrice     = "price"     // QueryPrice command for current price queries
	QueryRawPrices = "rawprices" // QueryRawPrices command for raw price queries
	QueryAssets    = "assets"    // QueryAssets command for assets query
)

// QueryRawPricesResp response to a rawprice query.
type QueryRawPricesResp []PostedPrice

// String implementation of fmt.Stringer.
func (n QueryRawPricesResp) String() string {
	strBuilder := strings.Builder{}
	for _, v := range n {
		strBuilder.WriteString(v.String() + "\n")
	}
	return strBuilder.String()
}

// QueryAssetsResp response to a assets query.
type QueryAssetsResp []string

// String implementation of fmt.Stringer.
func (n QueryAssetsResp) String() string {
	return strings.Join(n[:], "\n")
}
