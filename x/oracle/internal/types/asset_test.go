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

	// check invalid assetCode 1
	{
		a := NewAsset("Upper-case", oracles, true)
		require.Error(t, a.ValidateBasic())
	}

	// check invalid assetCode 2
	{
		a := NewAsset("non_ascii_ẞ_symbol", oracles, true)
		require.Error(t, a.ValidateBasic())
	}

	// check no oracles
	{
		a := NewAsset("dn2eth", Oracles{}, true)
		require.Error(t, a.ValidateBasic())
	}

	// check valid assetCode
	{
		a := NewAsset("dn2eth", oracles, true)
		require.NoError(t, a.ValidateBasic())
	}
}
