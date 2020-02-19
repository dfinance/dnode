package queries

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/WingsDao/wings-blockchain/x/currencies/types"
)

func TestQueryDestroyRes_String(t *testing.T) {
	t.Parallel()

	target := QueryDestroyRes{
		Destroy: types.Destroy{
			ID:        sdk.Int{},
			ChainID:   "",
			Symbol:    "",
			Amount:    sdk.Int{},
			Spender:   nil,
			Recipient: "",
			Timestamp: 0,
			TxHash:    "",
		},
	}
	require.Equal(t, target.Destroy.String(), target.String())
}

func TestQueryDestroysRes_String(t *testing.T) {
	t.Parallel()
	target := QueryDestroysRes{{
		Destroy: types.Destroy{
			ID:        sdk.Int{},
			ChainID:   "",
			Symbol:    "",
			Amount:    sdk.Int{},
			Spender:   nil,
			Recipient: "",
			Timestamp: 0,
			TxHash:    "",
		},
	}}
	require.Equal(t, target[0].Destroy.String(), target.String())
}
