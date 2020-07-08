// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Check MsgSubmitCall ValidateBasic.
func TestMSMsg_SubmitCall_ValidateBasic(t *testing.T) {
	t.Parallel()

	// ok
	{
		msg := NewOkMockMsMsg()
		target := NewMsgSubmitCall(msg, "unique", sdk.AccAddress("addr1"))
		require.NoError(t, target.ValidateBasic())
	}

	// fail: empty uniqueID
	{
		msg := NewOkMockMsMsg()
		target := NewMsgSubmitCall(msg, "", sdk.AccAddress("addr1"))
		require.Error(t, target.ValidateBasic())
	}

	// fail: empty creator
	{
		msg := NewOkMockMsMsg()
		target := NewMsgSubmitCall(msg, "unique", sdk.AccAddress{})
		require.Error(t, target.ValidateBasic())
	}

	// fail: invalid msg
	{
		msg := NewInvalidMockMsMsg()
		target := NewMsgSubmitCall(msg, "unique", sdk.AccAddress{})
		require.Error(t, target.ValidateBasic())
	}
}

// Test MsgSubmitCall implements sdk.Msg interface.
func TestMSMsg_SubmitCall_MsgInterface(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress("addr1")
	target := NewMsgSubmitCall(NewOkMockMsMsg(), "unique", addr)
	require.Equal(t, "submit_call", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{addr}, target.GetSigners())
}
