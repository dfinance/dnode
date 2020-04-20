package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	PriceBytesLimit = 8
)

// implement fmt.Stringer
func (a PendingPriceAsset) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s`, a.AssetCode))
}

// PendingPriceAsset struct that contains the info about the asset which price is still to be determined
type PendingPriceAsset struct {
	AssetCode string `json:"asset_code"`
}

func ValidateAddress(address string) (sdk.AccAddress, error) {
	pricefeed, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	return pricefeed, nil
}

func ParsePricefeeds(addresses string) (PriceFeeds, error) {
	res := make([]PriceFeed, 0)
	for _, address := range strings.Split(addresses, ",") {
		address = strings.TrimSpace(address)
		if len(address) == 0 {
			continue
		}
		pricefeedAddress, err := ValidateAddress(address)
		if err != nil {
			return nil, err
		}

		pricefeed := NewOracle(pricefeedAddress)

		res = append(res, pricefeed)
	}

	return res, nil
}
