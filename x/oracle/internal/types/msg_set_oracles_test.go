// +build unit

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOracleMsg_SetOracle(t *testing.T) {
	t.Parallel()

	nominee := sdk.AccAddress([]byte("someName"))
	oracle := NewOracle(sdk.AccAddress([]byte("oracle")))
	oracles := []Oracle{oracle}
	denom := "uftm"

	t.Run("GetSign", func(t *testing.T) {
		target := NewMsgSetOracles(nominee, denom, oracles)
		require.Equal(t, "set_oracles", target.Type())
		require.Equal(t, RouterKey, target.Route())
		require.True(t, len(target.GetSignBytes()) > 0)
		require.Equal(t, []sdk.AccAddress{nominee}, target.GetSigners())
	})

	t.Run("ValidateBasic", func(t *testing.T) {
		// ok
		{
			msg := NewMsgSetOracles(nominee, denom, oracles)
			require.NoError(t, msg.ValidateBasic())
		}

		// fail: invalid denom
		{
			msg := NewMsgSetOracles(nominee, "", oracles)
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid oracles
		{
			msg := NewMsgSetOracles(nominee, denom, []Oracle{})
			require.Error(t, msg.ValidateBasic())
		}

		// fail: invalid nominee
		{
			msg := NewMsgSetOracles(sdk.AccAddress{}, denom, oracles)
			require.Error(t, msg.ValidateBasic())
		}
	})

}
