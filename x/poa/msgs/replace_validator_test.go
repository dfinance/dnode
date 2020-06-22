// +build unit

package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
	poatypes "github.com/dfinance/dnode/x/poa/types"
)

func Test_MsgReplaceValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgReplaceValidator(sdkAddress, sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty old validator sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgReplaceValidator([]byte{}, sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty new validator empty sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgReplaceValidator(sdkAddress, []byte{}, ethAddress, sdkAddress).ValidateBasic())
	// empty new validator empty sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgReplaceValidator(sdkAddress, []byte{}, "", sdkAddress).ValidateBasic())
	// empty new validator ethAddress
	utils.CheckExpectedErr(t, poatypes.ErrWrongEthereumAddress, NewMsgReplaceValidator(sdkAddress, sdkAddress, "not_empty", sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgReplaceValidator(sdkAddress, sdkAddress, ethAddress, []byte{}).ValidateBasic())
}
