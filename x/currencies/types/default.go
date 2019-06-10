package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
)

const (
	ModuleName 	  				      	  = "currencies"

	DefaultRoute 	  					  = ModuleName
	DefaultCodespace  sdk.CodespaceType   = ModuleName
	DefaultParamspace 					  = ModuleName
)

var (
	DenomListKey = []byte("denoms")
)

// Key for storing currency
func GetCurrencyKey(symbol string) []byte {
	return []byte(fmt.Sprintf("currency:%s", symbol))
}

type Denoms []string