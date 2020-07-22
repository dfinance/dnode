// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/helpers/tests/utils"
	"github.com/dfinance/dnode/x/common_vm"
)

// Test MsgExecuteScript.
func TestVM_MsgExecuteScript(t *testing.T) {
	t.Parallel()

	acc := sdk.AccAddress([]byte("addr1"))
	code := make(Contract, 128)

	args := []ScriptArg{
		{Type: vm_grpc.VMTypeTag_U64, Value: []byte{0x1, 0x2, 0x3, 0x4}},
		{Type: vm_grpc.VMTypeTag_Vector, Value: []byte{0x0}},
		{Type: vm_grpc.VMTypeTag_Address, Value: common_vm.Bech32ToLibra(acc)},
	}

	msg := NewMsgExecuteScript(acc, code, args)
	require.Equal(t, msg.Signer, acc)
	require.Equal(t, msg.Script, code)
	require.NoError(t, msg.ValidateBasic())
	require.Equal(t, RouterKey, msg.Route())
	require.Equal(t, MsgExecuteScriptType, msg.Type())
	require.Equal(t, msg.GetSigners(), []sdk.AccAddress{acc})
	require.Equal(t, getMsgSignBytes(t, msg), msg.GetSignBytes())

	require.EqualValues(t, msg.Args, args)

	// message without signer
	msg = NewMsgExecuteScript([]byte{}, code, nil)
	require.Empty(t, msg.Signer)
	require.Nil(t, msg.Args)
	utils.CheckExpectedErr(t, sdkErrors.ErrInvalidAddress, msg.ValidateBasic())

	// message without args should be fine
	msg = NewMsgExecuteScript(acc, code, nil)
	require.NoError(t, msg.ValidateBasic())

	// script without code
	msg = NewMsgExecuteScript(acc, []byte{}, nil)
	utils.CheckExpectedErr(t, ErrEmptyContract, msg.ValidateBasic())
}

// Test new argument
func TestVM_NewScriptArg(t *testing.T) {
	t.Parallel()

	value := []byte{0, 1}
	tagType := vm_grpc.VMTypeTag_U64
	arg := NewScriptArg(tagType, value)
	require.Equal(t, tagType, arg.Type)
	require.Equal(t, value, arg.Value)
}
