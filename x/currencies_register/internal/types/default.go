package types

import (
	"fmt"
)

const (
	ModuleName = "currencies_register"
	StoreKey   = ModuleName
	RouterKey  = "curencyinfo"
)

var (
	DefaultOwner = make([]byte, 24)
)

func GetCurrencyPathKey(denom string) []byte {
	return []byte(fmt.Sprintf("currency_path:%s", denom))
}
