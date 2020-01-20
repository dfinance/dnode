package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func getMsgSignBytes(t *testing.T, msg sdk.Msg) []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	return sdk.MustSortJSON(bc)
}

func TestMsgDeployContract(t *testing.T) {
	t.Parallel()

	acc := sdk.AccAddress([]byte("addr1"))
	code := make(Contract, 128)
	msg := NewMsgDeployContract(acc, code)

	require.Equal(t, msg.Signer, acc)
	require.Equal(t, msg.Contract, code)
	require.NoError(t, msg.ValidateBasic())
	require.Equal(t, RouterKey, msg.Route())
	require.Equal(t, msg.GetSigners(), []sdk.AccAddress{acc})
	require.Equal(t, getMsgSignBytes(t, msg), msg.GetSignBytes())

	msg = NewMsgDeployContract([]byte{}, code)
	require.Empty(t, msg.Signer)

	err := msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeInvalidAddress)
	require.Equal(t, err.Codespace(), sdk.CodespaceRoot)

	msg = NewMsgDeployContract(acc, Contract{})
	require.Empty(t, msg.Contract)

	err = msg.ValidateBasic()
	require.Error(t, err)
	require.Equal(t, err.Code(), sdk.CodeType(CodeEmptyContractCode))
	require.Equal(t, err.Codespace(), Codespace)
}
