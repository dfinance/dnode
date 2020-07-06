// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPOA_Msg_AddValidator_Validate(t *testing.T) {
	t.Parallel()

	sdkAddr := sdk.AccAddress("addr1")
	ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"

	// ok
	{
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgAddValidator(sdkAddr, ethAddr, senderAddr)
		require.NoError(t, msg.ValidateBasic())
	}

	// fail: empty sender
	{
		senderAddr := sdk.AccAddress("")
		msg := NewMsgAddValidator(sdkAddr, ethAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}
}
