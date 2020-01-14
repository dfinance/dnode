// Default types for currency module implementation.
package types

import (
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName                         = "currencies"
	Router                             = ModuleName
	RouterKey                          = ModuleName
	DefaultCodespace sdk.CodespaceType = ModuleName
)

var (
	KeyDelimiter = []byte(":")
	DestroyQueue = []byte("destroy")
)

// Key for storing currency
func GetCurrencyKey(symbol string) []byte {
	return []byte(fmt.Sprintf("currency:%s", symbol))
}

// Key for issues
func GetIssuesKey(issueID string) []byte {
	return []byte(fmt.Sprintf("issues:%s", issueID))
}

// Get destroy key
func GetDestroyKey(id sdk.Int) []byte {
	return bytes.Join(
		[][]byte{
			DestroyQueue,
			[]byte(id.String()),
		},
		KeyDelimiter,
	)
}

// Get last ID key
func GetLastIDKey() []byte {
	return []byte("lastID")
}
