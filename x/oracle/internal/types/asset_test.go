// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func Test_NewAsset(t *testing.T) {
	t.Parallel()

	oracleAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	oracles := Oracles{Oracle{Address: oracleAddr}}

	// check invalid assetCode 1 (non lower-cased)
	{
		a := NewAsset("EthBtc", oracles, true)
		require.Error(t, a.ValidateBasic())
	}

	// check invalid assetCode 2 (non-ASCII symbol)
	{
		a := NewAsset("aẞb", oracles, true)
		require.Error(t, a.ValidateBasic())
	}

	// check invalid assetCode 3 (non-letter symbol)
	{
		a := NewAsset("a_b1", oracles, true)
		require.Error(t, a.ValidateBasic())
	}

	// check no oracles
	{
		a := NewAsset("dn_eth", Oracles{}, true)
		require.Error(t, a.ValidateBasic())
	}

	// check valid assetCode
	{
		a := NewAsset("dn_eth", oracles, true)
		require.NoError(t, a.ValidateBasic())
	}
}
