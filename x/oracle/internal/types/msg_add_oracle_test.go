// +build unit

package types

import (
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Check MsgAddOracle validate basic.
func TestOracleMsg_AddOracle(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	oracle := sdk.AccAddress([]byte("someOracle"))
	assetCode := dnTypes.AssetCode("btn_dfi")

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgAddOracle(nominee, assetCode, oracle)
		require.Equal(t, "add_oracle", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// ok
		{
			msg := NewMsgAddOracle(nominee, assetCode, oracle)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid assetCode
		{
			msg := NewMsgAddOracle(nominee, "", oracle)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid oracle
		{
			msg := NewMsgAddOracle(nominee, assetCode, sdk.AccAddress{})
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgAddOracle(sdk.AccAddress{}, assetCode, oracle)
			require.Error(t, msg.ValidateBasic())
		}
	})
}
