// +build unit

package msgs

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
)

type InvalidMsg struct{}

func (InvalidMsg) Route() string { return "" }
func (InvalidMsg) Type() string  { return "" }
func (InvalidMsg) ValidateBasic() error {
	return fmt.Errorf("some error")
}

func Test_SubmitCallValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")
	okMsg := NewMsgConfirmCall(0, sdkAddress)
	invalidMsg := InvalidMsg{}

	// correct
	require.Nil(t, NewMsgSubmitCall(okMsg, "", sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, NewMsgSubmitCall(okMsg, "", []byte{}).ValidateBasic())
	// invalid msg
	require.NotNil(t, NewMsgSubmitCall(invalidMsg, "", []byte{}).ValidateBasic())
}
