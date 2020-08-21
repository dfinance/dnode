// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMarkets_MsgCreateMarket_Valid(t *testing.T) {
	t.Parallel()

	addr := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h")

	msg := NewMsgCreateMarket(addr, "btc", "xfi")
	require.NoError(t, msg.ValidateBasic())
}

func TestMarkets_MsgCreateMarket_Invalid(t *testing.T) {
	t.Parallel()

	// empty from
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "btc", "xfi")
		require.Error(t, msg.ValidateBasic())

	}

	// empty baseDenom
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "", "xfi")
		require.Error(t, msg.ValidateBasic())

	}

	// empty quoteDenom
	{
		msg := NewMsgCreateMarket(sdk.AccAddress{}, "btc", "")
		require.Error(t, msg.ValidateBasic())

	}
}
