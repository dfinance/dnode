// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestOracleMsg_AddOracle(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	oracle := sdk.AccAddress([]byte("someOracle"))
	denom := "uftm"

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgAddOracle(nominee, denom, oracle)
		require.Equal(t, "add_oracle", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// ok
		{
			msg := NewMsgAddOracle(nominee, denom, oracle)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid denom
		{
			msg := NewMsgAddOracle(nominee, "", oracle)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid oracle
		{
			msg := NewMsgAddOracle(nominee, denom, sdk.AccAddress{})
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgAddOracle(sdk.AccAddress{}, denom, oracle)
			require.Error(t, msg.ValidateBasic())
		}
	})
}
