// +build unit

package vmauth

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

const (
	accBytes1 = "01000000030000006466690100000000000000000000000000000000000000000000002800000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000028000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	toDecode  = "000000000000000000000000000000000000000000000000000000000000000000000000"
)

func TestMarshalEmpty(t *testing.T) {
	accRes := AccountResource{
		WithdrawEvents: &EventHandle{},
		DepositEvents:  &EventHandle{},
	}
	AccResToBytes(accRes)
}

func TestUnmarshalEmpty(t *testing.T) {
	bz, err := hex.DecodeString(toDecode)
	if err != nil {
		t.Fatal(err)
	}

	BytesToAccRes(bz)
}

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
	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
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
		WithdrawEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		DepositEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		EventGenerator: 0,
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
		WithdrawEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		DepositEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		EventGenerator: 0,
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

	accRes := AccResFromAccount(&acc, nil)

	for i, coin := range acc.Coins {
		require.EqualValues(t, coin.Denom, accRes.Balances[i].Denom)
		require.EqualValues(t, coin.Amount, accRes.Balances[i].Value)
	}
}

func TestAccResFromSource(t *testing.T) {
	source := AccountResource{
		Balances: []DNCoin{
			{
				Denom: []byte("mmm"),
				Value: sdk.NewInt(1),
			},
		},
		WithdrawEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		DepositEvents: &EventHandle{
			Count: 0,
			Key:   make([]byte, 40),
		},
		EventGenerator: 1,
	}

	acc := auth.NewBaseAccountWithAddress(sdk.AccAddress("tmp"))
	if err := acc.SetCoins(sdk.Coins{sdk.NewCoin("dfi", sdk.NewInt(1))}); err != nil {
		t.Fatal(err)
	}

	accRes := AccResFromAccount(&acc, &source)

	for i, coin := range acc.Coins {
		require.EqualValues(t, coin.Denom, accRes.Balances[i].Denom)
		require.EqualValues(t, coin.Amount, accRes.Balances[i].Value)
	}

	require.Equal(t, source.EventGenerator, accRes.EventGenerator)
	require.EqualValues(t, source.DepositEvents, accRes.DepositEvents)
	require.EqualValues(t, source.WithdrawEvents, accRes.WithdrawEvents)
}

func TestGetResPath(t *testing.T) {
	res, err := hex.DecodeString(resourceKey)
	if err != nil {
		t.Fatal(err)
	}

	require.EqualValues(t, res, GetResPath())
}
