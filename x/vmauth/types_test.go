package vmauth

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers"
)

const (
	accBytes1 = "010000000300000064666901000000000000000000000000000000"
)

func TestBalancesToCoins(t *testing.T) {
	wbCoins := []DNCoin{
		{
			Denom: []byte("dfi"),
			Value: sdk.NewInt(1),
		},
		{
			Denom: []byte("eth"),
			Value: sdk.NewInt(1),
		},
	}

	coins := balancesToCoins(wbCoins)
	for i, coin := range coins {
		require.EqualValues(t, coin.Denom, wbCoins[i].Denom)
		require.EqualValues(t, coin.Amount, wbCoins[i].Value)
	}

	// check nil.
	coins = balancesToCoins(nil)
	require.Empty(t, coins)
}

func TestAddrToPathAddr(t *testing.T) {
	addr, err := sdk.AccAddressFromBech32("cosmos14ng6lzsvyy26sxmujmjthvrjde8x6gkk2gzeft")
	if err != nil {
		helpers.CrashWithError(err)
	}

	libraAddr := AddrToPathAddr(addr)

	config := sdk.GetConfig()
	prefix := config.GetBech32AccountAddrPrefix()
	zeros := make([]byte, libraAddressLength-len(prefix)-len(addr))

	bytes := make([]byte, 0)
	bytes = append(bytes, []byte(prefix)...)
	bytes = append(bytes, zeros...)
	bytes = append(bytes, addr...)

	require.EqualValues(t, bytes, libraAddr)
}

func TestBytesToAccRes(t *testing.T) {
	acc := AccountResource{
		Balances: []DNCoin{
			{
				Denom: []byte("dfi"),
				Value: sdk.NewInt(1),
			},
		},
	}

	bz := AccResToBytes(acc)

	newAcc := BytesToAccRes(bz)

	require.EqualValues(t, acc, newAcc)
}

func TestAccToBytes(t *testing.T) {
	acc := AccountResource{
		Balances: []DNCoin{
			{
				Denom: []byte("dfi"),
				Value: sdk.NewInt(1),
			},
		},
	}

	bz := AccResToBytes(acc)
	res, err := hex.DecodeString(accBytes1)
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, res, bz)
}

func TestAccResourceFromAccount(t *testing.T) {
	acc := auth.NewBaseAccountWithAddress(sdk.AccAddress("tmp"))
	if err := acc.SetCoins(sdk.Coins{sdk.NewCoin("dfi", sdk.NewInt(1))}); err != nil {
		t.Fatal(err)
	}

	accRes := AccResFromAccount(&acc)

	for i, coin := range acc.Coins {
		require.EqualValues(t, coin.Denom, accRes.Balances[i].Denom)
		require.EqualValues(t, coin.Amount, accRes.Balances[i].Value)
	}
}

func TestGetResPath(t *testing.T) {
	res, err := hex.DecodeString(resourceKey)
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, res, GetResPath())
}
