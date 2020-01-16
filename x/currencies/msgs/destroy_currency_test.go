package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"wings-blockchain/x/currencies/types"
)

func TestMsgDestroyCurrency_ValidateBasic(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("chainID", "symbol", sdk.NewInt(10), sdk.AccAddress([]byte("addr1")), "recipient")
	require.NoError(t, target.ValidateBasic())

	invalidTarget := target
	invalidTarget.Spender = sdk.AccAddress([]byte{})
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Recipient = ""
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Symbol = ""
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Amount = sdk.NewInt(0)
	require.Error(t, invalidTarget.ValidateBasic())
}

func TestMsgDestroyCurrency_Type(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("chainID", "symbol", sdk.NewInt(10), sdk.AccAddress([]byte("addr1")), "recipient")
	require.Equal(t, "destroy_currency", target.Type())
}

func TestMsgDestroyCurrency_Route(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("chainID", "symbol", sdk.NewInt(10), sdk.AccAddress([]byte("addr1")), "recipient")
	require.Equal(t, types.RouterKey, target.Route())
}

func TestMsgDestroyCurrency_GetSignBytes(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("chainID", "symbol", sdk.NewInt(10), sdk.AccAddress([]byte("addr1")), "recipient")
	require.True(t, len(target.GetSignBytes()) > 0)
}

func TestMsgDestroyCurrency_GetSigners(t *testing.T) {
	t.Parallel()

	target := NewMsgDestroyCurrency("chainID", "symbol", sdk.NewInt(10), sdk.AccAddress([]byte("addr1")), "recipient")
	require.Equal(t, []sdk.AccAddress{target.Spender}, target.GetSigners())
}
