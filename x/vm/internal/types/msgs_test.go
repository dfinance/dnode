package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

func getMsgSignBytes(t *testing.T, msg sdk.Msg) []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	return sdk.MustSortJSON(bc)
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

	err := msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeInvalidAddress)
	require.Equal(t, err.Codespace(), sdk.CodespaceRoot)

	msg = NewMsgDeployModule(acc, Contract{})
	require.Empty(t, msg.Module)

	err = msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeType(CodeEmptyContractCode))
	require.Equal(t, err.Codespace(), Codespace)
}

// Test MsgExecuteScript.
func TestMsgExecuteScript(t *testing.T) {
	t.Parallel()

	acc := sdk.AccAddress([]byte("addr1"))
	code := make(Contract, 128)

	args := make([]ScriptArg, 3)
	args[0] = NewScriptArg("10", vm_grpc.VMTypeTag_U64)
	args[1] = NewScriptArg("0x00", vm_grpc.VMTypeTag_ByteArray)
	args[2] = NewScriptArg(acc.String(), vm_grpc.VMTypeTag_Address)

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

	err := msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeInvalidAddress)
	require.Equal(t, err.Codespace(), sdk.CodespaceRoot)

	// message without args should be fine
	msg = NewMsgExecuteScript(acc, code, nil)
	require.NoError(t, msg.ValidateBasic())

	// script without code
	msg = NewMsgExecuteScript(acc, []byte{}, nil)
	err = msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeType(CodeEmptyContractCode))
	require.Equal(t, err.Codespace(), Codespace)
}

// Test new argument
func TestNewScriptArg(t *testing.T) {
	t.Parallel()

	value := "100"
	tagType := vm_grpc.VMTypeTag_U64
	arg := NewScriptArg(value, tagType)
	require.Equal(t, tagType, arg.Type)
	require.Equal(t, value, arg.Value)
}
