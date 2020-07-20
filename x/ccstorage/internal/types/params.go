package types

import (
	"encoding/hex"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store key
var (
	ParamStoreKeyCurrencies = []byte("currencies")
)

// CurrencyParams defines currency genesis params.
type CurrencyParams struct {
	// Currency decimals count
	Decimals uint8 `json:"decimals" yaml:"decimals"`
	// Path used to store account balance for currency denom (0x1::Dfinance::T<Coin>)
	BalancePathHex string `json:"balance_path_hex" yaml:"balance_path_hex"`
	// Path used to store CurrencyInfo for currency denom (0x1::Dfinance::Info<Coin>)
	InfoPathHex string `json:"info_path_hex" yaml:"info_path_hex"`
}

// CurrenciesParams is a map with denom key and CurrencyParams value, used for parameters storage.
type CurrenciesParams map[string]CurrencyParams

// Validate check that params are valid.
func (p CurrencyParams) Validate() error {
	if len(p.BalancePathHex) == 0 {
		return fmt.Errorf("balancePathHex: empty")
	}
	if len(p.InfoPathHex) == 0 {
		return fmt.Errorf("infoPathHex: empty")
	}
	if _, err := hex.DecodeString(p.BalancePathHex); err != nil {
		return fmt.Errorf("balancePathHex: %w", err)
	}
	if _, err := hex.DecodeString(p.InfoPathHex); err != nil {
		return fmt.Errorf("infoPathHex: %w", err)
	}

	return nil
}

// BalancePath return []byte representation for BalancePathHex.
func (p CurrencyParams) BalancePath() []byte {
	bytes, _ := hex.DecodeString(p.BalancePathHex)
	return bytes
}

// BalancePath return []byte representation for InfoPathHex.
func (p CurrencyParams) InfoPath() []byte {
	bytes, _ := hex.DecodeString(p.InfoPathHex)
	return bytes
}

// NewCurrencyParams creates a new CurrencyParams object.
func NewCurrencyParams(decimals uint8, balancePath, currencyInfoPath []byte) CurrencyParams {
	return CurrencyParams{
		Decimals:       decimals,
		BalancePathHex: hex.EncodeToString(balancePath),
		InfoPathHex:    hex.EncodeToString(currencyInfoPath),
	}
}

// ParamKeyTable returns Key declaration for parameters storage.
func ParamKeyTable() params.KeyTable {
	nilValidator := func(value interface{}) error { return nil }

	return params.NewKeyTable(
		params.NewParamSetPair(ParamStoreKeyCurrencies, CurrenciesParams{}, nilValidator),
	)
}
