// +build unit

package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests"
)

func Test_MsgRemoveValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgRemoveValidator(sdkAddress, sdkAddress).ValidateBasic())
	// invalid sdkAddress
	tests.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgRemoveValidator([]byte{}, sdkAddress).ValidateBasic())
	// invalid sender
	tests.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgRemoveValidator(sdkAddress, []byte{}).ValidateBasic())
}
