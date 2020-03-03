package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Asset struct that represents an asset in the oracle
type Asset struct {
	AssetCode  string  `json:"asset_code" yaml:"asset_code"`
	BaseAsset  string  `json:"base_asset" yaml:"base_asset"`
	QuoteAsset string  `json:"quote_asset" yaml:"quote_asset"`
	Oracles    Oracles `json:"oracles" yaml:"oracles"`
	Active     bool    `json:"active" yaml:"active"`
}

// NewAsset creates a new asset
func NewAsset(
	assetCode, baseAsset, quoteAsset string,
	oracles Oracles,
	active bool,
) Asset {
	return Asset{
		AssetCode:  assetCode,
		BaseAsset:  baseAsset,
		QuoteAsset: quoteAsset,
		Oracles:    oracles,
		Active:     active,
	}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (a Asset) ValidateBasic() sdk.Error {
	if err := assetCodeFilter(a.AssetCode); err != nil {
		return sdk.ErrInternal(fmt.Sprintf("invalid assetCode: Value: %s. Error: %v", a.AssetCode, err))
	}

	if len(a.BaseAsset) == 0 {
		return sdk.ErrInternal(fmt.Sprintf("invalid TokenRecord: BaseAsset: %s. Error: Missing BaseAsset", a.BaseAsset))
	}
	if len(a.QuoteAsset) == 0 {
		return sdk.ErrInternal(fmt.Sprintf("invalid TokenRecord: QuoteAsset: %s. Error: Missing QuoteAsset", a.QuoteAsset))
	}
	if len(a.Oracles) == 0 {
		return sdk.ErrInternal("invalid TokenRecord: Error: Missing Oracles")
	}
	return nil
}

// implement fmt.Stringer
func (a Asset) String() string {
	return fmt.Sprintf(`Asset:
	Asset Code: %s
	Base Asset: %s
	Quote Asset: %s
	Oracles: %s
	Active: %t`,
		a.AssetCode, a.BaseAsset, a.QuoteAsset, a.Oracles, a.Active)
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
	AssetCode string  `json:"asset_code" yaml:"asset_code"`
	Price     sdk.Int `json:"price" yaml:"price"`
}

// PostedPrice struct represented a price for an asset posted by a specific oracle
type PostedPrice struct {
	AssetCode     string         `json:"asset_code" yaml:"asset_code"`
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address"`
	Price         sdk.Int        `json:"price" yaml:"price"`
	Expiry        time.Time      `json:"expiry" yaml:"expiry"`
}

// implement fmt.Stringer
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
Price: %s`, cp.AssetCode, cp.Price))
}

// implement fmt.Stringer
func (pp PostedPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf(`AssetCode: %s
OracleAddress: %s
Price: %s
Expiry: %s`, pp.AssetCode, pp.OracleAddress, pp.Price, pp.Expiry))
}

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int           { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j]) }
