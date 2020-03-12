package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_MsgRemoveValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgRemoveValidator(sdkAddress, sdkAddress).ValidateBasic())
	// invalid sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgRemoveValidator([]byte{}, sdkAddress).ValidateBasic())
	// invalid sender
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgRemoveValidator(sdkAddress, []byte{}).ValidateBasic())
}