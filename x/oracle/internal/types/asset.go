package types

import (
	"fmt"
	"strings"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Asset struct that represents an asset in the oracle.
type Asset struct {
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code" example:"btc_dfi"`
	// List of registered RawPrice sources
	Oracles Oracles `json:"oracles" yaml:"oracles"`
	// Not used ATM
	Active bool `json:"active" yaml:"active"`
}

func (a Asset) String() string {
	return fmt.Sprintf("Asset:\n"+
		"  AssetCode: %s\n"+
		"  Oracles: %s\n"+
		"  Active: %v",
		a.AssetCode, a.Oracles, a.Active)
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (a Asset) ValidateBasic() error {
	if err := a.AssetCode.Validate(); err != nil {
		return sdkErrors.Wrapf(ErrInternal, "invalid assetCode: value (%s), error (%v)", a.AssetCode, err)
	}

	if len(a.Oracles) == 0 {
		return sdkErrors.Wrap(ErrInternal, "invalid TokenRecord: missing Oracles")
	}

	return nil
}

// NewAsset creates a new asset
func NewAsset(assetCode dnTypes.AssetCode, oracles Oracles, active bool) Asset {
	return Asset{
		AssetCode: assetCode,
		Oracles:   oracles,
		Active:    active,
	}
}

// Assets slice type for oracle.
type Assets []Asset

func (list Assets) String() string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString("Assets:\n")
	for i, asset := range list {
		strBuilder.WriteString(asset.String())
		if i < len(list) - 1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}
