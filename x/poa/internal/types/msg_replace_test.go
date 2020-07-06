// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPOA_Msg_ReplaceValidator_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		sdkAddrOld := sdk.AccAddress("addr1")
		sdkAddrNew := sdk.AccAddress("addr2")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgReplaceValidator(sdkAddrOld, sdkAddrNew, ethAddr, senderAddr)
		require.NoError(t, msg.ValidateBasic())
	}

	// fail: empty oldValidator
	{
		sdkAddrOld := sdk.AccAddress("")
		sdkAddrNew := sdk.AccAddress("addr2")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgReplaceValidator(sdkAddrOld, sdkAddrNew, ethAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}

	// fail: equal addresses
	{
		sdkAddrOld := sdk.AccAddress("addr2")
		sdkAddrNew := sdk.AccAddress("addr2")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		senderAddr := sdk.AccAddress("sender1")
		msg := NewMsgReplaceValidator(sdkAddrOld, sdkAddrNew, ethAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}

	// fail: empty sender
	{
		sdkAddrOld := sdk.AccAddress("addr1")
		sdkAddrNew := sdk.AccAddress("addr2")
		ethAddr := "0x6adaF04f4E2BA9CDdE3ec143bdcF02AD830c1b7d"
		senderAddr := sdk.AccAddress("")
		msg := NewMsgReplaceValidator(sdkAddrOld, sdkAddrNew, ethAddr, senderAddr)
		require.Error(t, msg.ValidateBasic())
	}
}
