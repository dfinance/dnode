// +build unit

package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/poa/types"
)

func Test_MsgAddValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")

	// correct
	require.Nil(t, NewMsgAddValidator(sdkAddress, ethAddress, sdkAddress).ValidateBasic())
	// empty validator sdkAddress
	tests.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgAddValidator([]byte{}, ethAddress, sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	tests.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgAddValidator(sdkAddress, ethAddress, []byte{}).ValidateBasic())
	// empty validator ethAddress
	tests.CheckExpectedErr(t, types.ErrWrongEthereumAddress, NewMsgAddValidator(sdkAddress, "", sdkAddress).ValidateBasic())
	// invalid validator ethAddress
	tests.CheckExpectedErr(t, types.ErrWrongEthereumAddress, NewMsgAddValidator(sdkAddress, "not_empty", sdkAddress).ValidateBasic())
}
