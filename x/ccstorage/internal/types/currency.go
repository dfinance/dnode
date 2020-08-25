package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/glav"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Currency is an info object with currency params.
type Currency struct {
	// Currency denom (symbol)
	Denom string `json:"denom" yaml:"denom" example:"xfi"`
	// Number of currency decimals
	Decimals uint8 `json:"decimals" yaml:"decimals" example:"0"`
	// Total amount of currency coins in Bank
	Supply sdk.Int `json:"supply" yaml:"supply" swaggertype:"string" example:"100"`
}

// Valid checks that Currency is valid.
func (c Currency) Valid() error {
	if err := dnTypes.DenomFilter(c.Denom); err != nil {
		return fmt.Errorf("denom is invalid: %v", err)
	}

	return nil
}

// GetSupplyCoin creates sdk.Coin with supply amount.
func (c Currency) GetSupplyCoin() sdk.Coin {
	return sdk.NewCoin(c.Denom, c.Supply)
}

// BalancePath return []byte representation for BalancePath.
func (c Currency) BalancePath() []byte {
	return glav.BalanceVector(c.Denom)
}

// BalancePathHex return string representation for InfoPath.
func (c Currency) BalancePathHex() string {
	return hex.EncodeToString(c.BalancePath())
}

// InfoPath return []byte representation for InfoPath.
func (c Currency) InfoPath() []byte {
	return glav.CurrencyInfoVector(c.Denom)
}

// InfoPathHex return string representation for InfoPath.
func (c Currency) InfoPathHex() string {
	return hex.EncodeToString(c.BalancePath())
}

// UintToDec converts sdk.Uint to sdk.Dec using currency decimals.
func (c Currency) UintToDec(quantity sdk.Uint) sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.Int(quantity), int64(c.Decimals))
}

// DecToUint converts sdk.Dec to sdk.Uint using currency decimals.
func (c Currency) DecToUint(quantity sdk.Dec) sdk.Uint {
	res := quantity.Quo(c.MinDecimal()).TruncateInt()

	return sdk.NewUintFromBigInt(res.BigInt())
}

// MinDecimal return minimal currency value.
func (c Currency) MinDecimal() sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.OneInt(), int64(c.Decimals))
}

func (c Currency) String() string {
	return fmt.Sprintf("Currency:\n"+
		"  Denom:    %s\n"+
		"  Decimals: %d\n"+
		"  Supply:   %s",
		c.Denom,
		c.Decimals,
		c.Supply.String(),
	)
}

// Currencies is a slice of Currency objects.
type Currencies []Currency

// ToParams converts Currencies to CurrenciesParams.
func (list Currencies) ToParams() CurrenciesParams {
	var params CurrenciesParams
	for _, currency := range list {
		params = append(params, CurrencyParams{
			Denom:    currency.Denom,
			Decimals: currency.Decimals,
		})
	}

	return params
}

// NewCurrency creates a new Currency object.
func NewCurrency(params CurrencyParams, supply sdk.Int) Currency {
	return Currency{
		Denom:    params.Denom,
		Decimals: params.Decimals,
		Supply:   supply,
	}
}
