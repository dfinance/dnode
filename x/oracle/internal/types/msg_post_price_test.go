// +build unit

package types

import (
	"math/big"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestOracleMsg_PostPrice(t *testing.T) {
	t.Parallel()

	from := sdk.AccAddress([]byte("someName"))
	assetCode := "btc_dfi"
	price := sdk.NewInt(30050000)
	expiry := time.Now()
	negativePrice, _ := sdk.NewIntFromString("-1")
	bigInt := sdk.NewIntFromBigInt(big.NewInt(0).SetBit(big.NewInt(0), PriceBytesLimit*8, 1))

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgPostPrice(from, assetCode, price, expiry)
		require.Equal(t, "post_price", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{from}, target.GetSigners())
	})

	t.Run("GetSign", func(t *testing.T) {
		// ok
		{
			msg := NewMsgPostPrice(from, assetCode, price, expiry)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid from
		{
			msg := NewMsgPostPrice(sdk.AccAddress{}, assetCode, price, expiry)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid assetCode
		{
			msg := NewMsgPostPrice(from, "", price, expiry)
			require.Error(t, msg.ValidateBasic())
		}
		// fail: invalid price over limit
		{
			msg := NewMsgPostPrice(from, assetCode, bigInt, expiry)
			require.Error(t, msg.ValidateBasic())
		}
		// fail: invalid price negative
		{
			msg := NewMsgPostPrice(from, assetCode, negativePrice, expiry)
			require.Error(t, msg.ValidateBasic())
		}
	})
}
