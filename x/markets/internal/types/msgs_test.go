// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_MsgCreateMarket_Valid(t *testing.T) {
	addr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")

	msg := NewMsgCreateMarket(addr, "btc", "dfi")
	require.NoError(t, msg.ValidateBasic())
}

func Test_MsgCreateMarket_Invalid(t *testing.T) {
	// empty from
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "btc", "dfi")
		require.Error(t, msg.ValidateBasic())

	}

	// empty baseDenom
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "", "dfi")
		require.Error(t, msg.ValidateBasic())

	}

	// empty quoteDenom
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "btc", "")
		require.Error(t, msg.ValidateBasic())

	}
}
