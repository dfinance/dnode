package msgs

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/currencies/types"
)

func TestMsgIssueCurrency_ValidateBasic(t *testing.T) {
	t.Parallel()

	target := NewMsgIssueCurrency("symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")), "issue1")
	require.NoError(t, target.ValidateBasic())

	invalidTarget := target
	invalidTarget.Recipient = sdk.AccAddress([]byte{})
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Symbol = ""
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Decimals = -1
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.Amount = sdk.NewInt(0)
	require.Error(t, invalidTarget.ValidateBasic())

	invalidTarget = target
	invalidTarget.IssueID = ""
	require.Error(t, invalidTarget.ValidateBasic())
}

func TestMsgIssueCurrency_Route(t *testing.T) {
	t.Parallel()

	target := NewMsgIssueCurrency("symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")), "issue1")
	require.Equal(t, types.RouterKey, target.Route())
}

func TestMsgIssueCurrency_Type(t *testing.T) {
	t.Parallel()

	target := NewMsgIssueCurrency("symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")), "issue1")
	require.Equal(t, "issue_currency", target.Type())
}

func TestMsgIssueCurrency_GetSignBytes(t *testing.T) {
	t.Parallel()

	target := NewMsgIssueCurrency("symbol", sdk.NewInt(10), 0, sdk.AccAddress([]byte("addr1")), "issue1")
	require.True(t, len(target.GetSignBytes()) > 0)

}

func TestMsgIssueCurrency_GetSigners(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress([]byte("addr1"))
	target := NewMsgIssueCurrency("symbol", sdk.NewInt(10), 0, addr, "issue1")
	require.Equal(t, []sdk.AccAddress{}, target.GetSigners())
	require.Equal(t, 0, len(target.GetSigners()))
}
