package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
)

// MarketExtended is a Market object extended with currency info from currencies module.
// Object is used by orders module.
type MarketExtended struct {
	// Market unique ID
	ID dnTypes.ID `json:"id" yaml:"id" format:"string representation for big.Uint" swaggertype:"string" example:"0"`
	// Base asset currency (for ex. btc)
	BaseCurrency ccstorage.Currency `json:"base_currency" yaml:"base_currency"`
	// Quote asset currency (for ex. dfi)
	QuoteCurrency ccstorage.Currency `json:"quote_currency" yaml:"quote_currency"`
}

// BaseToQuoteQuantity converts base asset price and quantity to quote asset quantity.
// Function normalizes quantity to be used later by OrderBook module, that way quantity for bid and ask
// order are casted to the same base (base quantity).
func (m MarketExtended) BaseToQuoteQuantity(basePrice sdk.Uint, baseQuantity sdk.Uint) (sdk.Uint, error) {
	// convert basePrice (in QuoteAsset denom) to sdk.Dec
	basePriceDec := m.QuoteCurrency.UintToDec(basePrice)
	// convert baseQuantity (amount for BaseAsset) to sdk.Dec
	baseQuantityDec := m.BaseCurrency.UintToDec(baseQuantity)

	// get QuoteAsset quantity in sdk.Dec
	quoteQuantityDec := basePriceDec.Mul(baseQuantityDec)

	// check if result is lower than the lowest quote volume
	if quoteQuantityDec.LT(m.QuoteCurrency.MinDecimal()) {
		return sdk.Uint{}, sdkErrors.Wrapf(
			ErrInvalidQuantity,
			"quantity is too small (basePrice: %s -> %s, baseQuantity: %s -> %s): %s LT %s",
			basePrice,
			basePriceDec,
			baseQuantity,
			baseQuantityDec,
			quoteQuantityDec,
			m.QuoteCurrency.MinDecimal(),
		)
	}

	// convert result to sdk.Uint (remove decimal part)
	quoteQuantity := m.QuoteCurrency.DecToUint(quoteQuantityDec)

	return quoteQuantity, nil
}

// BaseDenom return string base asset denom representation.
func (m MarketExtended) BaseDenom() string {
	return string(m.BaseCurrency.Denom)
}

// BaseDenom return string quote asset denom representation.
func (m MarketExtended) QuoteDenom() string {
	return string(m.QuoteCurrency.Denom)
}

// String returns multi-line text object representation.
func (m MarketExtended) String() string {
	b := strings.Builder{}
	b.WriteString("MarketExtended:\n")
	b.WriteString(fmt.Sprintf("  ID: %s\n", m.ID.String()))
	b.WriteString(fmt.Sprintf("  BaseCurrency: %s\n", m.BaseCurrency.String()))
	b.WriteString(fmt.Sprintf("  QuoteCurrency: %s\n", m.QuoteCurrency.String()))

	return b.String()
}

// TableHeaders returns table headers for multi-line text table object representation.
func (m MarketExtended) TableHeaders() []string {
	return []string{
		"M.ID",
		"M.BaseAssetDenom",
		"M.BaseAssetDecimals",
		"M.QuoteAssetDenom",
		"M.QuoteAssetDecimals",
	}
}

// TableHeaders returns table rows for multi-line text table object representation.
func (m MarketExtended) TableValues() []string {
	return []string{
		m.ID.String(),
		m.BaseDenom(),
		strconv.FormatUint(uint64(m.BaseCurrency.Decimals), 10),
		m.QuoteDenom(),
		strconv.FormatUint(uint64(m.QuoteCurrency.Decimals), 10),
	}
}

func NewMarketExtended(market Market, baseCurrency, quoteCurrency ccstorage.Currency) MarketExtended {
	return MarketExtended{
		ID:            market.ID,
		BaseCurrency:  baseCurrency,
		QuoteCurrency: quoteCurrency,
	}
}
