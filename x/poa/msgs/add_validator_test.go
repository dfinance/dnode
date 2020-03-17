package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/poa/types"
)

func Test_MsgAddValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgAddValidator(sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty validator sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgAddValidator([]byte{}, ethAddress, sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgAddValidator(sdkAddress, ethAddress, []byte{}).ValidateBasic())
	// empty validator ethAddress
	checkExpectedErr(t, sdk.ErrUnknownRequest(""), NewMsgAddValidator(sdkAddress, "", sdkAddress).ValidateBasic())
	// invalid validator ethAddress
	checkExpectedErr(t, types.ErrWrongEthereumAddress(""), NewMsgAddValidator(sdkAddress, "not_empty", sdkAddress).ValidateBasic())
}
