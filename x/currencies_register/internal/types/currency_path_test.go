// +build unit

package types

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// New currency path test.
func TestNewCurrencyPath(t *testing.T) {
	path := make([]byte, 32)

	currPath := NewCurrencyPath(path)
	require.EqualValues(t, path, currPath.Path)
}

// Test currency path to string.
func TestCurrencyPath_String(t *testing.T) {
	path := make([]byte, 32)

	currPath := NewCurrencyPath(path)
	currPathStr := fmt.Sprintf("Path: %s", hex.EncodeToString(currPath.Path))
	require.Equal(t, currPathStr, currPath.String())
}
