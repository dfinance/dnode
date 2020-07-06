// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPOA_Msg_RemoveValidator_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		sdkAddr := sdk.AccAddress("addr1")
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgRemoveValidator(sdkAddr, senderAddr)
		require.NoError(t, msg.ValidateBasic())
	}

	// fail: empty address
	{
		sdkAddr := sdk.AccAddress("")
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgRemoveValidator(sdkAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}

	// fail: empty sender
	{
		sdkAddr := sdk.AccAddress("addr1")
		senderAddr := sdk.AccAddress("")
		msg := NewMsgRemoveValidator(sdkAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}
}
