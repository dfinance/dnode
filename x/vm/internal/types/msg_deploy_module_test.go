// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
)

func getMsgSignBytes(t *testing.T, msg sdk.Msg) []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Test MsgDeployModule.
func TestMsgDeployModule(t *testing.T) {
	t.Parallel()

	acc := sdk.AccAddress([]byte("addr1"))
	code := make(Contract, 128)
	msg := NewMsgDeployModule(acc, code)

	require.Equal(t, msg.Signer, acc)
	require.Equal(t, msg.Module, code)
	require.NoError(t, msg.ValidateBasic())
	require.Equal(t, RouterKey, msg.Route())
	require.Equal(t, MsgDeployModuleType, msg.Type())
	require.Equal(t, msg.GetSigners(), []sdk.AccAddress{acc})
	require.Equal(t, getMsgSignBytes(t, msg), msg.GetSignBytes())

	msg = NewMsgDeployModule([]byte{}, code)
	require.Empty(t, msg.Signer)
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, msg.ValidateBasic())

	msg = NewMsgDeployModule(acc, Contract{})
	require.Empty(t, msg.Module)
	utils.CheckExpectedErr(t, ErrEmptyContract, msg.ValidateBasic())
}
