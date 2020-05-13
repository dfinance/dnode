// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StringFilter(t *testing.T) {
	// ASCII string
	require.NoError(t, StringFilter("abc_123DEF", nil, []RuneFilterOpt{RuneIsASCII}))

	// string with non ASCII symbol
	require.Error(t, StringFilter("abc_®_123DEF", nil, []RuneFilterOpt{RuneIsASCII}))

	// lower case string
	require.NoError(t, StringFilter("abc_123def", nil, []RuneFilterOpt{RuneLetterIsLowerCase}))

	// string with non lower case symbol
	require.Error(t, StringFilter("abc_123Def", nil, []RuneFilterOpt{RuneLetterIsLowerCase}))

	// ASCII and lower case combined check 1
	require.Error(t, StringFilter("abc_®123Def", nil, []RuneFilterOpt{RuneIsASCII, RuneLetterIsLowerCase}))

	// ASCII and lower case combined check 2
	require.Error(t, StringFilter("abc_123Def_®", nil, []RuneFilterOpt{RuneIsASCII, RuneLetterIsLowerCase}))

	// empty string
	require.Error(t, StringFilter("", []StrFilterOpt{StringIsEmpty}, nil))
}
