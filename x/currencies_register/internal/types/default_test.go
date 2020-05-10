// +build unit

package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Get currency path.
func TestGetCurrencyPathKey(t *testing.T) {
	denom := "dfi"
	storagePath := []byte(fmt.Sprintf("currency_path:%s", denom))

	require.EqualValues(t, storagePath, GetCurrencyPathKey(denom))
}
