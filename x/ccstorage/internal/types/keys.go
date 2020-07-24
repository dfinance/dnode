package types

import "bytes"

var (
	KeyDelimiter      = []byte(":")
	KeyCurrencyPrefix = []byte("currency")
)

// GetCurrencyKey returns Key for storing currency.
func GetCurrencyKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			KeyCurrencyPrefix,
			[]byte(denom),
		},
		KeyDelimiter,
	)
}

// GetCurrencyKeyPrefix return currency storage key prefix (used for iteration).
func GetCurrencyKeyPrefix() []byte {
	return append(KeyCurrencyPrefix, KeyDelimiter...)
}

// GetCurrencyBalancePathKey returns storage key for currencyBalance VM path.
func GetCurrencyBalancePathKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("currencyBalancePath"),
			[]byte(denom),
		},
		KeyDelimiter,
	)
}

// GetCurrencyInfoPathKey returns storage key for currencyInfo VM path.
func GetCurrencyInfoPathKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("currencyInfoPath"),
			[]byte(denom),
		},
		KeyDelimiter,
	)
}
