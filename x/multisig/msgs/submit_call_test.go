package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type InvalidMsg struct{}

func (InvalidMsg) Route() string { return "" }
func (InvalidMsg) Type() string  { return "" }
func (InvalidMsg) ValidateBasic() sdk.Error {
	return sdk.NewError(sdk.CodespaceType(0), 0, "some error")
}

func Test_SubmitCallValidator(t *testing.T) {
	t.Parallel()

	sdkAddress, _ := sdk.AccAddressFromHex("0102030405060708090A0102030405060708090A")
	okMsg := NewMsgConfirmCall(0, sdkAddress)
	invalidMsg := InvalidMsg{}

	// correct
	require.Nil(t, NewMsgSubmitCall(okMsg, "", sdkAddress).ValidateBasic())
	// empty sender sdkAddress
	checkExpectedErr(t, sdk.ErrInvalidAddress(""), NewMsgSubmitCall(okMsg, "", []byte{}).ValidateBasic())
	// invalid msg
	require.NotNil(t, NewMsgSubmitCall(invalidMsg, "", []byte{}).ValidateBasic())
}
