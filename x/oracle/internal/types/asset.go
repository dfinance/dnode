package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Asset struct that represents an asset in the oracle
type Asset struct {
	AssetCode string  `json:"asset_code" yaml:"asset_code" example:"dfi"`
	Oracles   Oracles `json:"oracles" yaml:"oracles"` // List of registered RawPrice sources
	Active    bool    `json:"active" yaml:"active"`   // Not used ATM
}

// NewAsset creates a new asset
func NewAsset(
	assetCode string,
	oracles Oracles,
	active bool,
) Asset {
	return Asset{
		AssetCode: assetCode,
		Oracles:   oracles,
		Active:    active,
	}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (a Asset) ValidateBasic() error {
	if err := dnTypes.AssetCodeFilter(a.AssetCode); err != nil {
		return sdkErrors.Wrapf(ErrInternal, "invalid assetCode: value (%s), error (%v)", a.AssetCode, err)
	}

	if len(a.Oracles) == 0 {
		return sdkErrors.Wrap(ErrInternal, "invalid TokenRecord: missing Oracles")
	}

	return nil
}

// implement fmt.Stringer
func (a Asset) String() string {
	return fmt.Sprintf(`Asset:
	Asset Code: %s
	Oracles: %s
	Active: %t`,
		a.AssetCode, a.Oracles, a.Active)
}

// Assets array type for oracle
type Assets []Asset

// String implements fmt.Stringer
func (as Assets) String() string {
	out := "Assets:\n"
	for _, a := range as {
		out += fmt.Sprintf("%s\n", a.String())
	}

	return strings.TrimSpace(out)
}

// Oracle struct that documents which address an oracle is using
type Oracle struct {
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// String implements fmt.Stringer
func (o Oracle) String() string {
	return fmt.Sprintf(`Address: %s`, o.Address)
}

func NewOracle(address sdk.AccAddress) Oracle {
	return Oracle{
		Address: address,
	}
}

// Oracles array type for oracle
type Oracles []Oracle

// String implements fmt.Stringer
func (os Oracles) String() string {
	out := "Oracles:\n"
	for _, o := range os {
		out += fmt.Sprintf("%s\n", o.String())
	}

	return strings.TrimSpace(out)
}

// CurrentPrice struct that contains the metadata of a current price for a particular asset in the oracle module.
type CurrentPrice struct {
	AssetCode  string    `json:"asset_code" yaml:"asset_code" example:"dfi"` // Denom
	Price      sdk.Int   `json:"price" yaml:"price" swaggertype:"string" example:"1000"`
	ReceivedAt time.Time `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"` // Timestamp Price createdAt
}

// PostedPrice struct represented a price for an asset posted by a specific oracle
type PostedPrice struct {
	AssetCode     string         `json:"asset_code" yaml:"asset_code" example:"dfi"`                                                                        // Denom
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address" swaggertype:"string" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"` // Price source
	Price         sdk.Int        `json:"price" yaml:"price" swaggertype:"string" example:"1000"`
	ReceivedAt    time.Time      `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"` // Timestamp Price createdAt
}

// implement fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
Price: %s
ReceivedAt: %s`, cp.AssetCode, cp.Price, cp.ReceivedAt))
}

// implement fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
OracleAddress: %s
Price: %s
ReceivedAt: %s`, pp.AssetCode, pp.OracleAddress, pp.Price, pp.ReceivedAt))
}

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int           { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j]) }
