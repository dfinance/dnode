package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ConfirmCallValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgConfirmCall(0, sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgConfirmCall(0, []byte{}).ValidateBasic())
}