// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_stringFilter(t *testing.T) {
	// ASCII string
	require.NoError(t, stringFilter("abc_123DEF", nil, []runeFilterOpt{runeIsASCII}))

	// string with non ASCII symbol
	require.Error(t, stringFilter("abc_®_123DEF", nil, []runeFilterOpt{runeIsASCII}))

	// lower case string
	require.NoError(t, stringFilter("abc_123def", nil, []runeFilterOpt{runeLetterIsLowerCase}))

	// string with non lower case symbol
	require.Error(t, stringFilter("abc_123Def", nil, []runeFilterOpt{runeLetterIsLowerCase}))

	// ASCII and lower case combined check 1
	require.Error(t, stringFilter("abc_®123Def", nil, []runeFilterOpt{runeIsASCII, runeLetterIsLowerCase}))

	// ASCII and lower case combined check 2
	require.Error(t, stringFilter("abc_123Def_®", nil, []runeFilterOpt{runeIsASCII, runeLetterIsLowerCase}))

	// empty string
	require.Error(t, stringFilter("", []strFilterOpt{stringIsEmpty}, nil))
}
