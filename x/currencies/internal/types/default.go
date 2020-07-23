package types

import (
	"bytes"
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName   = "currencies"
	RouterKey    = ModuleName
	StoreKey     = ModuleName
	GovRouterKey = RouterKey
)

var (
	IssuePrefix    = []byte("issue")
	WithdrawPrefix = []byte("withdraw")
	KeyDelimiter   = []byte(":")
)

// GetIssuesKey returns key for storing issues.
func GetIssuesKey(id string) []byte {
	return bytes.Join(
		[][]byte{
			IssuePrefix,
			[]byte(id),
		},
		KeyDelimiter,
	)
}

// GetIssuesPrefix returns key prefix for issue objects iteration.
func GetIssuesPrefix() []byte {
	return append(IssuePrefix, KeyDelimiter...)
}

// GetWithdrawsPrefix returns key prefix for withdraw objects iteration.
func GetWithdrawsPrefix() []byte {
	return append(WithdrawPrefix, KeyDelimiter...)
}

// MustParseIssueKey parses issue storage key.
func MustParseIssueKey(key []byte) string {
	values := bytes.Split(key, KeyDelimiter)
	if len(values) != 2 {
		panic(fmt.Errorf("key %q: invalid splitted length %d", string(key), len(values)))
	}

	if !bytes.Equal(values[0], IssuePrefix) {
		panic(fmt.Errorf("key %q: value[0] %q: wrong prefix", string(key), string(values[0])))
	}

	return string(values[1])
}

// GetWithdrawKey returns key for storing withdraw.
func GetWithdrawKey(id dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			WithdrawPrefix,
			[]byte(id.String()),
		},
		KeyDelimiter,
	)
}

// GetLastWithdrawIDKey returns storage key for withdrawID.
func GetLastWithdrawIDKey() []byte {
	return []byte("lastWithdrawID")
}
