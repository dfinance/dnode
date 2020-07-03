// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Check MsgRevokeConfirm ValidateBasic.
func TestMSMsg_RevokeConfirm_ValidateBasic(t *testing.T) {
	t.Parallel()

	// ok
	{
		msg := NewMsgRevokeConfirm(dnTypes.NewIDFromUint64(0), sdk.AccAddress("addr1"))
		require.NoError(t, msg.ValidateBasic())
	}

	// fail: id
	{
		msg := NewMsgRevokeConfirm(dnTypes.ID{}, sdk.AccAddress("addr1"))
		require.Error(t, msg.ValidateBasic())
	}

	// fail: sender
	{
		msg := NewMsgRevokeConfirm(dnTypes.NewIDFromUint64(0), sdk.AccAddress{})
		require.Error(t, msg.ValidateBasic())
	}
}

// Test MsgRevokeConfirm implements sdk.Msg interface.
func TestMSMsg_RevokeConfirm_MsgInterface(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress("addr1")
	target := NewMsgRevokeConfirm(dnTypes.NewIDFromUint64(0), addr)
	require.Equal(t, "revoke_confirm", target.Type())
	require.Equal(t, RouterKey, target.Route())
	require.True(t, len(target.GetSignBytes()) > 0)
	require.Equal(t, []sdk.AccAddress{addr}, target.GetSigners())
}
