package types

import (
	"fmt"
	"strings"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Asset struct that represents an asset in the oracle.
type Asset struct {
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code" example:"dfi"`
	Oracles   Oracles           `json:"oracles" yaml:"oracles"` // List of registered RawPrice sources
	Active    bool              `json:"active" yaml:"active"`   // Not used ATM
}

// String implement fmt.Stringer for the Asset type.
func (a Asset) String() string {
	return fmt.Sprintf(`Asset:
	Asset Code: %s
	Oracles: %s
	Active: %t`,
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
func NewAsset(
	assetCode dnTypes.AssetCode,
	oracles Oracles,
	active bool,
) Asset {
	return Asset{
		AssetCode: assetCode,
		Oracles:   oracles,
		Active:    active,
	}
}

// Assets array type for oracle.
type Assets []Asset

// String implements fmt.Stringer for the Assets type.
func (as Assets) String() string {
	out := "Assets:\n"
	for _, a := range as {
		out += fmt.Sprintf("%s\n", a.String())
	}

	return strings.TrimSpace(out)
}
