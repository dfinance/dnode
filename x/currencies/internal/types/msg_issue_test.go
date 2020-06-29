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

	target := NewMsgIssueCurrency("issue1", "symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")))
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
		invalidTarget.Denom = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount (zero)
	{
		invalidTarget := target
		invalidTarget.Amount = sdk.NewInt(0)
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount (negative)
	{
		invalidTarget := target
		invalidTarget.Amount = sdk.NewInt(-1)
		require.Panics(t, func() { invalidTarget.ValidateBasic() })
	}

	// invalid: payee
	{
		invalidTarget := target
		invalidTarget.Payee = sdk.AccAddress([]byte{})
		require.Error(t, invalidTarget.ValidateBasic())
	}
}

// Test MsgIssueCurrency sdk.Msg params.
func TestCurrenciesMsg_IssueCurrency_MsgInterface(t *testing.T) {
	t.Parallel()

	target := NewMsgIssueCurrency("issue1", "symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")))
	require.Equal(t, "issue_currency", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{}, target.GetSigners())
	require.Equal(t, 0, len(target.GetSigners()))
}
