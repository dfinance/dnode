// +build integ

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
	require.Error(t, stringFilter("", []strFilterOpt{stringNotEmpty}, nil))

	// delimiter: ok
	require.NoError(t, stringFilter("abc_def", []strFilterOpt{newDelimiterStrFilterOpt("_")}, nil))

	// delimiter: none
	require.Error(t, stringFilter("abcdef", []strFilterOpt{newDelimiterStrFilterOpt("_")}, nil))

	// delimiter: is prefix
	require.Error(t, stringFilter("_abcdef", []strFilterOpt{newDelimiterStrFilterOpt("_")}, nil))

	// delimiter: is suffix
	require.Error(t, stringFilter("abcdef_", []strFilterOpt{newDelimiterStrFilterOpt("_")}, nil))

	// delimiter: multiple
	require.Error(t, stringFilter("abc_d_ef", []strFilterOpt{newDelimiterStrFilterOpt("_")}, nil))
}

func Test_assetCodeFilter(t *testing.T) {
	// ok
	require.NoError(t, AssetCodeFilter("eth_usdt"))

	// fail: empty
	require.Error(t, AssetCodeFilter(""))

	// fail: non lower cased letter
	require.Error(t, AssetCodeFilter("ETH_usdt"))

	// fail: invalid separator
	require.Error(t, AssetCodeFilter("eth:usdt"))

	// fail: invalid separator
	require.Error(t, AssetCodeFilter("eth__usdt"))

	// fail: non ASCII symbol
	require.Error(t, AssetCodeFilter("eth®_usdt"))

	// fail: non letter symbol
	require.Error(t, AssetCodeFilter("eth_usdt1"))
}

func Test_denomFilter(t *testing.T) {
	// ok
	require.NoError(t, DenomFilter("dfi"))

	// fail: empty
	require.Error(t, DenomFilter(""))

	// fail: non lower cased letter
	require.Error(t, DenomFilter("Dfi"))

	// fail: non ASCII symbol
	require.Error(t, DenomFilter("eth®"))

	// fail: non letter symbol
	require.Error(t, DenomFilter("eth1"))
}
