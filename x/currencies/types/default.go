package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
)

const (
	ModuleName 	  				      	  = "currencies"

	DefaultRoute 	  					  = ModuleName
	DefaultCodespace  sdk.CodespaceType   = ModuleName
)

// Key for storing currency
func GetCurrencyKey(symbol string) []byte {
	return []byte(fmt.Sprintf("currency:%s", symbol))
}

// Key for issues
func GetIssuesKey(issueID string) []byte {
	return []byte(fmt.Sprintf("issues:%s", issueID))
}
