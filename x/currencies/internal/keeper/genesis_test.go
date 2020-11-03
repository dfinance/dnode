//+build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

func TestCurrenciesKeeper_Genesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	lastID := dnTypes.NewIDFromUint64(2)
	state := types.GenesisState{
		Issues: []types.GenesisIssue{
			{
				Issue: types.NewIssue(
					sdk.NewCoin("xfi", sdk.NewInt(150)),
					sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				),
				ID: "issue1",
			},
			{
				Issue: types.NewIssue(
					sdk.NewCoin("eth", sdk.NewInt(250)),
					sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				),
				ID: "issue2",
			},
		},
		Withdraws: types.Withdraws{
			types.NewWithdraw(
				dnTypes.NewIDFromUint64(0),
				sdk.NewCoin("xfi", sdk.NewInt(100)),
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				"pgAcc1",
				"pgID",
				1,
				[]byte("hash1"),
			),
			types.NewWithdraw(
				dnTypes.NewIDFromUint64(1),
				sdk.NewCoin("eth", sdk.NewInt(200)),
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				"pgAcc2",
				"pgID",
				1,
				[]byte("hash2"),
			),
			types.NewWithdraw(
				dnTypes.NewIDFromUint64(2),
				sdk.NewCoin("btc", sdk.NewInt(300)),
				sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()),
				"pgAcc3",
				"pgID",
				1,
				[]byte("hash3"),
			),
		},
		LastWithdrawID: &lastID,
	}

	// init
	{
		keeper.InitGenesis(ctx, keeper.cdc.MustMarshalJSON(state))

		// lastID
		require.Equal(t, state.LastWithdrawID.String(), keeper.getLastWithdrawID(ctx).String())
		// issues
		require.Len(t, keeper.GetGenesisIssues(ctx), len(state.Issues))
		for i, getIssue := range keeper.GetGenesisIssues(ctx) {
			require.EqualValues(t, state.Issues[i], getIssue)
		}
		// withdraws
		require.Len(t, keeper.getWithdraws(ctx), len(state.Withdraws))
		for i, getWithdraw := range keeper.getWithdraws(ctx) {
			require.EqualValues(t, state.Withdraws[i], getWithdraw)
		}
	}

	// export
	{
		var state types.GenesisState
		keeper.cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &state)

		// lastID
		require.NotNil(t, state.LastWithdrawID)
		require.Equal(t, keeper.getLastWithdrawID(ctx).String(), state.LastWithdrawID.String())
		// issues
		require.Len(t, keeper.GetGenesisIssues(ctx), len(state.Issues))
		for i, getIssue := range keeper.GetGenesisIssues(ctx) {
			require.EqualValues(t, getIssue, state.Issues[i])
		}
		// withdraws
		require.Len(t, keeper.getWithdraws(ctx), len(state.Withdraws))
		for i, getWithdraw := range keeper.getWithdraws(ctx) {
			require.EqualValues(t, getWithdraw, state.Withdraws[i])
		}
	}
}
