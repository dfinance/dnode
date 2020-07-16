package types

import (
	"strings"
)

const (
	QueryPrice     = "price"
	QueryRawPrices = "rawprices"
	QueryAssets    = "assets"
)

// Client response for rawPrices request.
type QueryRawPricesResp []PostedPrice

func (n QueryRawPricesResp) String() string {
	strBuilder := strings.Builder{}
	for i, v := range n {
		strBuilder.WriteString(v.String())
		if i < len(n)-1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}

// Client response for all assets request.
type QueryAssetsResp []string

func (n QueryAssetsResp) String() string {
	return strings.Join(n[:], "\n")
}
