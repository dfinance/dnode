// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test MsgDestroyCurrency ValidateBasic.
func TestCurrenciesMsg_DestroyCurrency_ValidateBasic(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("symbol", sdk.NewInt(10), sdk.AccAddress("addr1"), "recipient", "chainID")
	// ok
	{
		require.NoError(t, target.ValidateBasic())
	}

	// invalid: denom
	{
		invalidTarget := target
		invalidTarget.Denom = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount
	{
		invalidTarget := target
		invalidTarget.Amount = sdk.NewInt(0)
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: spender
	{
		invalidTarget := target
		invalidTarget.Spender = sdk.AccAddress([]byte{})
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: recipient
	{
		invalidTarget := target
		invalidTarget.Recipient = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}
}

// Test MsgDestroyCurrency sdk.Msg params.
func TestCurrenciesMsg_DestroyCurrency_MsgInterface(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("symbol", sdk.NewInt(10), sdk.AccAddress("addr1"), "recipient", "chainID")
	require.Equal(t, "destroy_currency", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{target.Spender}, target.GetSigners())
}
