package types

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/olekukonko/tablewriter"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Market object.
type Market struct {
	// Market unique ID
	ID                dnTypes.ID `json:"id" yaml:"id"`
	// Base asset denomination (for ex. BTC)
	BaseAssetDenom    string     `json:"base_asset_denom" yaml:"base_asset_denom"`
	// Quote asset denomination (for ex. DFI)
	QuoteAssetDenom   string     `json:"quote_asset_denom" yaml:"quote_asset_denom"`
	// Number of decimals used for base asset quantity representation
	BaseAssetDecimals uint8      `json:"base_asset_decimals" yaml:"base_asset_decimals"`
}

// QuantityToDecimal converts quantity to sdk.Dec with Market specifics.
func (m Market) QuantityToDecimal(quantity sdk.Uint) sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.Int(quantity), int64(m.BaseAssetDecimals))
}

// BaseToQuoteQuantity converts base asset price and quantity to quote asset quantity.
// Function normalizes quantity to be used later by OrderBook module, that way quantity for bid and ask
// orders are casted to the same base (base quantity).
func (m Market) BaseToQuoteQuantity(basePrice sdk.Uint, baseQuantity sdk.Uint) (sdk.Uint, error) {
	pDec := sdk.NewDecFromBigInt(basePrice.BigInt())
	qDec := m.QuantityToDecimal(baseQuantity)

	resDec := pDec.Mul(qDec)
	if resDec.IsZero() {
		return sdk.Uint{}, sdkErrors.Wrap(ErrInvalidQuantity, "quantity is too small")
	}
	resUint := sdk.NewUintFromBigInt(resDec.TruncateInt().BigInt())

	return resUint, nil
}

// Valid check object validity.
func (m Market) Valid() error {
	if err := m.ID.Valid(); err != nil {
		return sdkErrors.Wrap(ErrWrongID, err.Error())
	}
	if m.BaseAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "BaseAsset")
	}
	if m.QuoteAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "QuoteAsset")
	}

	return nil
}

// Strings returns multi-line text object representation.
func (m Market) String() string {
	b := strings.Builder{}
	b.WriteString("Market:\n")
	b.WriteString(fmt.Sprintf("  ID:           %s\n", m.ID.String()))
	b.WriteString(fmt.Sprintf("  BaseAsset:    %s\n", m.BaseAssetDenom))
	b.WriteString(fmt.Sprintf("  QuoteAsset:   %s\n", m.QuoteAssetDenom))
	b.WriteString(fmt.Sprintf("  BaseDecimals: %d\n", m.BaseAssetDecimals))

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (m Market) TableHeaders() []string {
	return []string{
		"M.ID",
		"M.BaseAsset",
		"M.QuoteAsset",
		"M.BaseDecimals",
	}
}

// TableHeaders returns table rows for multi-line text table object representation.
func (m Market) TableValues() []string {
	return []string{
		m.ID.String(),
		m.BaseAssetDenom,
		m.QuoteAssetDenom,
		strconv.FormatUint(uint64(m.BaseAssetDecimals), 10),
	}
}

// NewMarket create a new market object.
func NewMarket(id dnTypes.ID, baseAsset, quoteAsset string, baseDecimals uint8) Market {
	return Market{
		ID:                id,
		BaseAssetDenom:    baseAsset,
		QuoteAssetDenom:   quoteAsset,
		BaseAssetDecimals: baseDecimals,
	}
}

// Market slice type.
type Markets []Market

// Strings returns multi-line text object representation.
func (l Markets) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader(Market{}.TableHeaders())

	for _, m := range l {
		t.Append(m.TableValues())
	}
	t.Render()

	return string(buf.Bytes())
}
