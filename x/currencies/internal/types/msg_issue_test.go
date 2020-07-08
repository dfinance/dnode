// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test MsgIssueCurrency ValidateBasic.
func TestCurrenciesMsg_IssueCurrency_ValidateBasic(t *testing.T) {
	t.Parallel()

	coin := sdk.NewCoin("symbol", sdk.NewInt(10))
	target := NewMsgIssueCurrency("issue1", coin, sdk.AccAddress([]byte("addr1")))
	// ok
	{
		require.NoError(t, target.ValidateBasic())
	}

	// invalid: id
	{
		invalidTarget := target
		invalidTarget.ID = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: denom
	{
		invalidTarget := target
		invalidTarget.Coin.Denom = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount (zero)
	{
		invalidTarget := target
		invalidTarget.Coin.Amount = sdk.NewInt(0)
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount (negative)
	{
		invalidTarget := target
		invalidTarget.Coin.Amount = sdk.NewInt(-1)
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: payee
	{
		invalidTarget := target
		invalidTarget.Payee = sdk.AccAddress([]byte{})
		require.Error(t, invalidTarget.ValidateBasic())
	}
}

// Test MsgIssueCurrency implements msmodule.MsMsg interface.
func TestCurrenciesMsg_IssueCurrency_MsgInterface(t *testing.T) {
	t.Parallel()

	coin := sdk.NewCoin("symbol", sdk.NewInt(10))
	target := NewMsgIssueCurrency("issue1", coin, sdk.AccAddress([]byte("addr1")))
	require.Equal(t, "issue_currency", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{}, target.GetSigners())
	require.Equal(t, 0, len(target.GetSigners()))
}
