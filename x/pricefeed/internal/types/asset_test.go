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

	pricefeedAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	pricefeeds := PriceFeeds{PriceFeed{Address: pricefeedAddr}}

	// check invalid assetCode 1
	{
		a := NewAsset("Upper-case", pricefeeds, true)
		require.Error(t, a.ValidateBasic())
	}

	// check invalid assetCode 2
	{
		a := NewAsset("non_ascii_ẞ_symbol", pricefeeds, true)
		require.Error(t, a.ValidateBasic())
	}

	// check no pricefeeds
	{
		a := NewAsset("dn2eth", PriceFeeds{}, true)
		require.Error(t, a.ValidateBasic())
	}

	// check valid assetCode
	{
		a := NewAsset("dn2eth", pricefeeds, true)
		require.NoError(t, a.ValidateBasic())
	}
}
