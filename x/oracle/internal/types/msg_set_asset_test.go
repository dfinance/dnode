// +build unit

package types

import (
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestOracleMsg_SetAsset(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	assetCode := dnTypes.AssetCode("btc_dfi")
	oracles := Oracles([]Oracle{NewOracle(sdk.AccAddress([]byte("someName")))})
	asset := NewAsset(assetCode, oracles, true)

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgSetAsset(nominee, asset)
		require.Equal(t, "set_asset", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("GetSign", func(t *testing.T) {
		// ok
		{
			msg := NewMsgSetAsset(nominee, asset)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid assetCode
		{
			tmpAsset := &asset
			asset2 := *tmpAsset
			asset2.AssetCode = dnTypes.AssetCode("wrong")
			msg := NewMsgAddAsset(nominee, asset2)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgSetAsset(sdk.AccAddress{}, asset)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid asset
		{
			msg := NewMsgSetAsset(nominee, Asset{})
			require.Error(t, msg.ValidateBasic())
		}
	})
}
