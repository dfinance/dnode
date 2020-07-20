// +build unit

package types

import (
	"encoding/hex"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

func Test_GetAccAddressFromHexOrBech32(t *testing.T) {
	bech32 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	addr, err := GetAccAddressFromHexOrBech32(bech32.String())
	require.NoError(t, err)

	require.Truef(t, bech32.Equals(addr), "bech32 addresses doesn't match", bech32, addr)

	libraAddr := hex.EncodeToString(bech32)

	addr, err = GetAccAddressFromHexOrBech32(libraAddr)
	require.NoError(t, err)

	require.Truef(t, addr.Equals(bech32), "libra hex addresses doesn't match", libraAddr, addr)
}
