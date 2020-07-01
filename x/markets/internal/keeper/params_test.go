// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

func TestMarketsKeeper_Params_StoreIO(t *testing.T) {
	input := NewTestInput(t)

	inParams := types.Params{
		Markets: types.Markets{
			types.Market{
				ID:              dnTypes.NewIDFromUint64(0),
				BaseAssetDenom:  "btc",
				QuoteAssetDenom: "dfi",
			},
			types.Market{
				ID:              dnTypes.NewIDFromUint64(1),
				BaseAssetDenom:  "eth",
				QuoteAssetDenom: "dfi",
			},
		},
	}
	input.keeper.SetParams(input.ctx, inParams)

	outParams := input.keeper.GetParams(input.ctx)
	require.NotNil(t, outParams)
	require.NotNil(t, outParams.Markets)
	require.Len(t, outParams.Markets, 2)
	require.True(t, outParams.Markets[0].ID.Equal(inParams.Markets[0].ID))
	require.True(t, outParams.Markets[1].ID.Equal(inParams.Markets[1].ID))
	require.Equal(t, outParams.Markets[0].BaseAssetDenom, inParams.Markets[0].BaseAssetDenom)
	require.Equal(t, outParams.Markets[1].BaseAssetDenom, inParams.Markets[1].BaseAssetDenom)
	require.Equal(t, outParams.Markets[0].QuoteAssetDenom, inParams.Markets[0].QuoteAssetDenom)
	require.Equal(t, outParams.Markets[1].QuoteAssetDenom, inParams.Markets[1].QuoteAssetDenom)
}
