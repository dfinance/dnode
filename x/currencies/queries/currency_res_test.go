package queries

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/WingsDao/wings-blockchain/x/currencies/types"
)

func TestQueryCurrencyRes_String(t *testing.T) {
	t.Parallel()

	target := QueryCurrencyRes{
		Currency: types.NewCurrency("test", sdk.NewInt(1), 0),
	}
	require.Equal(t, target.Currency.String(), target.String())
}
