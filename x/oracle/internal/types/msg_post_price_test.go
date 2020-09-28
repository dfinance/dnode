// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Check MsgPostPrice validate basic.
func TestOracleMsg_PostPrice(t *testing.T) {
	t.Parallel()

	from := sdk.AccAddress([]byte("someName"))
	assetCode := dnTypes.AssetCode("btc_xfi")
	askPrice := sdk.NewDec(30050005)
	bidPrice := sdk.NewDec(30050000)
	expiry := time.Now()
	negativePrice := sdk.NewDec(-1)

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgPostPrice(from, assetCode, askPrice, bidPrice, expiry)
		require.Equal(t, "post_price", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{from}, target.GetSigners())
	})

	t.Run("GetSign", func(t *testing.T) {
		// ok
		{
			msg := NewMsgPostPrice(from, assetCode, askPrice, bidPrice, expiry)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid from
		{
			msg := NewMsgPostPrice(sdk.AccAddress{}, assetCode, askPrice, bidPrice, expiry)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid assetCode
		{
			msg := NewMsgPostPrice(from, "", askPrice, bidPrice, expiry)
			require.Error(t, msg.ValidateBasic())
		}
		// fail: invalid price negative
		{
			msg := NewMsgPostPrice(from, assetCode, negativePrice, negativePrice, expiry)
			require.Error(t, msg.ValidateBasic())
		}
	})
}
