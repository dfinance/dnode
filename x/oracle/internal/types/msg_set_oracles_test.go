// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Check MsgSetOracles validate basic.
func TestOracleMsg_SetOracle(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	oracle := NewOracle(sdk.AccAddress([]byte("oracle")))
	oracles := []Oracle{oracle}
	assetCode := dnTypes.AssetCode("btc_xfi")

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgSetOracles(nominee, assetCode, oracles)
		require.Equal(t, "set_oracles", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// ok
		{
			msg := NewMsgSetOracles(nominee, assetCode, oracles)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid assetCode
		{
			msg := NewMsgSetOracles(nominee, "", oracles)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid oracles
		{
			msg := NewMsgSetOracles(nominee, assetCode, []Oracle{})
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgSetOracles(sdk.AccAddress{}, assetCode, oracles)
			require.Error(t, msg.ValidateBasic())
		}
	})

}
