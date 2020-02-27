// +build oracle-integration

package api

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	mnemonic = "tiny clump grief head sleep eager follow castle twelve stock hamster spend trumpet clump license rude enough afraid faith poem steel sun misery differ"
	chainID  = "wings-testnet"
)

func Test_GetAccount(t *testing.T) {
	fees, err := sdk.ParseCoins("1wings")
	require.NoError(t, err)
	cl, err := NewClient(mnemonic, chainID, "127.0.0.1:1317", fees)
	require.NoError(t, err)

	acc, err := cl.getAccount()
	require.NoError(t, err)
	require.NotNil(t, acc)
}

func Test_PostPrice(t *testing.T) {
	fees, err := sdk.ParseCoins("1wings")
	require.NoError(t, err)
	cl, err := NewClient(mnemonic, chainID, "127.0.0.1:1317", fees)
	require.NoError(t, err)

	require.NoError(t, cl.PostPrice("ETH_USDT", "10000.02"))
}
