// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func TestMarkets_Genesis_Valid(t *testing.T) {
	t.Parallel()

	lastID := dnTypes.NewIDFromUint64(1)
	state := GenesisState{
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
		LastMarketID: &lastID,
	}
	require.NoError(t, state.Validate())
}

func TestMarkets_Genesis_Invalid(t *testing.T) {
	t.Parallel()

	// invalid ID
	{
		lastID := dnTypes.NewIDFromUint64(0)
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.ID(sdk.Uint{}),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
			},
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}

	// invalid baseDenom
	{
		lastID := dnTypes.NewIDFromUint64(0)
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "BTC",
					QuoteAssetDenom: "dfi",
				},
			},
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}

	// invalid quoteDenom
	{
		lastID := dnTypes.NewIDFromUint64(0)
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi_1",
				},
			},
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}

	// duplicate market
	{
		lastID := dnTypes.NewIDFromUint64(0)
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
				Market{
					ID:              dnTypes.NewIDFromUint64(1),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "eth",
				},
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "usdt",
				},
			},
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}

	// lastID nil with existing markets
	{
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
			},
		}
		require.Error(t, state.Validate())
	}

	// lastID not nil without existing markets
	{
		lastID := dnTypes.NewIDFromUint64(0)
		state := GenesisState{
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}

	// lastID doesn't match max market ID
	{
		lastID := dnTypes.NewIDFromUint64(1)
		state := GenesisState{
			Markets: Markets{
				Market{
					ID:              dnTypes.NewIDFromUint64(0),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "dfi",
				},
				Market{
					ID:              dnTypes.NewIDFromUint64(1),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "eth",
				},
				Market{
					ID:              dnTypes.NewIDFromUint64(2),
					BaseAssetDenom:  "btc",
					QuoteAssetDenom: "usdt",
				},
			},
			LastMarketID: &lastID,
		}
		require.Error(t, state.Validate())
	}
}
