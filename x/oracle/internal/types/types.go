package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	oracle, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}
	return oracle, nil
}

func ParseOracles(addresses string) (Oracles, error) {
	res := make([]Oracle, 0)
	for _, address := range strings.Split(addresses, ",") {
		address = strings.TrimSpace(address)
		if len(address) == 0 {
			continue
		}
		oracleAddress, err := ValidateAddress(address)
		if err != nil {
			return nil, err
		}

		oracle := NewOracle(oracleAddress)

		res = append(res, oracle)
	}

	return res, nil
}
