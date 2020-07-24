// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Test MsgWithdrawCurrency ValidateBasic.
func TestCurrenciesMsg_WithdrawCurrency_ValidateBasic(t *testing.T) {
	t.Parallel()

	coin := sdk.NewCoin("symbol", sdk.NewInt(10))
	target := NewMsgWithdrawCurrency(coin, sdk.AccAddress("addr1"), "recipient", "chainID")
	// ok
	{
		require.NoError(t, target.ValidateBasic())
	}

	// invalid: denom
	{
		invalidTarget := target
		invalidTarget.Coin.Denom = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: amount
	{
		invalidTarget := target
		invalidTarget.Coin.Amount = sdk.NewInt(0)
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: payer
	{
		invalidTarget := target
		invalidTarget.Payer = sdk.AccAddress([]byte{})
		require.Error(t, invalidTarget.ValidateBasic())
	}

	// invalid: pegZonePayee
	{
		invalidTarget := target
		invalidTarget.PegZonePayee = ""
		require.Error(t, invalidTarget.ValidateBasic())
	}
}

// Test MsgWithdrawCurrency implements sdk.Msg interface.
func TestCurrenciesMsg_WithdrawCurrency_MsgInterface(t *testing.T) {
	t.Parallel()

	coin := sdk.NewCoin("symbol", sdk.NewInt(10))
	target := NewMsgWithdrawCurrency(coin, sdk.AccAddress("addr1"), "recipient", "chainID")
	require.Equal(t, "withdraw_currency", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{target.Payer}, target.GetSigners())
}
