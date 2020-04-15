// +build unit

package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func TestMsgSort(t *testing.T) {
	from := sdk.AccAddress([]byte("someName"))
	price := sdk.NewInt(1)
	expiry := time.Now()

	msg := types.NewMsgPostPrice(from, "uftm", price, expiry)

	fee := auth.NewStdFee(200000, nil)
	stdTx := auth.NewStdTx([]sdk.Msg{msg}, fee, []auth.StdSignature{}, "")
	signBytes := auth.StdSignBytes("dn", 4, 1, stdTx.Fee, stdTx.Msgs, stdTx.Memo)

	t.Logf("%s", signBytes)
	signed := auth.StdSignBytes(
		"", 4, 1, auth.NewStdFee(200000, nil), []sdk.Msg{msg}, "",
	)
	t.Logf("%s", signed)
}

func TestMsgPlaceBid_ValidateBasic(t *testing.T) {
	addr := sdk.AccAddress([]byte("someName"))
	// oracles := []Oracle{Oracle{
	// 	OracleAddress: addr.String(),
	// }}
	price := sdk.NewInt(30050000)
	expiry := time.Now().Add(time.Hour * 2)
	//negativeExpiry := time.Now()
	negativePrice, _ := sdk.NewIntFromString("-1")

	tests := []struct {
		name       string
		msg        types.MsgPostPrice
		expectPass bool
	}{
		{"normal", types.MsgPostPrice{addr, "dn", price, expiry}, true},
		{"emptyAddr", types.MsgPostPrice{sdk.AccAddress{}, "dn", price, expiry}, false},
		{"emptyAsset", types.MsgPostPrice{addr, "", price, expiry}, false},
		{"negativePrice", types.MsgPostPrice{addr, "dn", negativePrice, expiry}, false},
		//{"negativeExpiry", types.MsgPostPrice{addr, "dn", price, negativeExpiry}, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expectPass {
				require.Nil(t, tc.msg.ValidateBasic())
			} else {
				require.NotNil(t, tc.msg.ValidateBasic())
			}
		})
	}
}
