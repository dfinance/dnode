// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Check MsgConfirmCall ValidateBasic.
func TestMSMsg_ConfirmCall_ValidateBasic(t *testing.T) {
	t.Parallel()

	// ok
	{
		msg := NewMsgConfirmCall(dnTypes.NewIDFromUint64(0), sdk.AccAddress("addr1"))
		require.NoError(t, msg.ValidateBasic())
	}

	// fail: invalid id
	{
		msg := NewMsgConfirmCall(dnTypes.ID{}, sdk.AccAddress("addr1"))
		require.Error(t, msg.ValidateBasic())
	}

	// fail: invalid sender
	{
		msg := NewMsgConfirmCall(dnTypes.NewIDFromUint64(0), sdk.AccAddress{})
		require.Error(t, msg.ValidateBasic())
	}
}

// Test MsgConfirmCall implements sdk.Msg interface.
func TestMSMsg_ConfirmCall_MsgInterface(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress("addr1")
	target := NewMsgConfirmCall(dnTypes.NewIDFromUint64(0), addr)
	require.Equal(t, "confirm_call", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{addr}, target.GetSigners())
}
