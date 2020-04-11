// +build unit

package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	poatypes "github.com/dfinance/dnode/x/poa/types"
)

func Test_MsgReplaceValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgReplaceValidator(sdkAddress, sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty old validator sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgReplaceValidator([]byte{}, sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty new validator empty sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgReplaceValidator(sdkAddress, []byte{}, ethAddress, sdkAddress).ValidateBasic())
	// empty new validator empty sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgReplaceValidator(sdkAddress, []byte{}, "", sdkAddress).ValidateBasic())
	// empty new validator ethAddress
	checkExpectedErr(t, poatypes.ErrWrongEthereumAddress(""), NewMsgReplaceValidator(sdkAddress, sdkAddress, "not_empty", sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgReplaceValidator(sdkAddress, sdkAddress, ethAddress, []byte{}).ValidateBasic())
}
