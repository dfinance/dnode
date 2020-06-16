// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func Test_Params_Valid(t *testing.T) {
	params := Params{
		Markets: Markets{
			Market{
				ID:              dnTypes.NewIDFromUint64(0),
				BaseAssetDenom:  "btc",
				QuoteAssetDenom: "dfi",
			},
			Market{
				ID:              dnTypes.NewIDFromUint64(1),
				BaseAssetDenom:  "eth",
				QuoteAssetDenom: "dfi",
			},
		},
	}
	require.NoError(t, params.Validate())
}

func Test_Params_Invalid(t *testing.T) {
	// invalid ID
	{
		params := Params{
			Markets: Markets{
				Market{
					ID:              dnTypes.ID(sdk.Uint{}),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
			},
		}
		require.Error(t, params.Validate())
	}

	// invalid baseDenom
	{
		params := Params{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "BTC",
					QuoteAssetDenom: "dfi",
				},
			},
		}
		require.Error(t, params.Validate())
	}

	// invalid quoteDenom
	{
		params := Params{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi_1",
				},
			},
		}
		require.Error(t, params.Validate())
	}

	// invalid ID order
	{
		params := Params{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(1),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
			},
		}
		require.Error(t, params.Validate())
	}
}
