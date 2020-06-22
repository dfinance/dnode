// +build unit

package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
)

func Test_ConfirmCallValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgConfirmCall(0, sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgConfirmCall(0, []byte{}).ValidateBasic())
}
