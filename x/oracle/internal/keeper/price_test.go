// +build unit

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func TestOracleKeeper_CheckPriceReceiveTime(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	receivedAtDiffDur := time.Duration(keeper.GetPostPriceParams(ctx).ReceivedAtDiffInS) * time.Second

	// check equal timestamps
	{
		require.Nil(t, keeper.CheckPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time))
	}

	// check timestamps within +-range
	{
		dur := receivedAtDiffDur / 2
		require.Nil(t, keeper.CheckPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(dur)))
		require.Nil(t, keeper.CheckPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(-dur)))
	}

	// check timestamps outside of +-range
	{
		dur := receivedAtDiffDur + 1*time.Second
		utils.CheckExpectedErr(t, types.ErrInvalidReceivedAt, keeper.CheckPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(dur)))
		utils.CheckExpectedErr(t, types.ErrInvalidReceivedAt, keeper.CheckPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(-dur)))
	}
}

func TestOracleKeeper_SetPrice(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx
	header := ctx.BlockHeader()

	// Set price by oracle 1
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[0], input.stdAssetCode,
			sdk.NewInt(33000000),
			header.Time)
		require.NoError(t, err)

		// Get raw prices
		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 1)
		require.Equal(t, rawPrices[0].Price.Equal(sdk.NewInt(33000000)), true)
	}

	// Set price by oracle 2
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[1], input.stdAssetCode,
			sdk.NewInt(35000000),
			header.Time)
		require.NoError(t, err)

		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 2)
		require.Equal(t, rawPrices[1].Price.Equal(sdk.NewInt(35000000)), true)
	}

	// Update Price by oracle 1
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[0], input.stdAssetCode,
			sdk.NewInt(37000000),
			header.Time)
		require.NoError(t, err)

		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, rawPrices[0].Price.Equal(sdk.NewInt(37000000)), true)
	}
}

func TestOracleKeeper_GetRawPrice(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx
	header := ctx.BlockHeader()

	// Set price by oracle 1
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[0], input.stdAssetCode,
			sdk.NewInt(33000000),
			header.Time)
		require.NoError(t, err)

		// Get raw prices
		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 1)
		require.Equal(t, rawPrices[0].Price.Equal(sdk.NewInt(33000000)), true)
	}
}

func TestOracleKeeper_CurrentPrice(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx
	header := ctx.BlockHeader()

	keeper.SetPrice(
		ctx, input.addresses[0], input.stdAssetCode,
		sdk.NewInt(33000000),
		header.Time)
	keeper.SetPrice(
		ctx, input.addresses[1], input.stdAssetCode,
		sdk.NewInt(35000000),
		header.Time)
	keeper.SetPrice(
		ctx, input.addresses[2], input.stdAssetCode,
		sdk.NewInt(34000000),
		header.Time)
	// Set current price
	err := keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	// Get Current price
	price := keeper.GetCurrentPrice(ctx, input.stdAssetCode)
	require.Equal(t, price.Price.Equal(sdk.NewInt(34000000)), true)

	// Even number of oracles
	keeper.SetPrice(
		ctx, input.addresses[0], input.stdAssetCode,
		sdk.NewInt(33000000),
		header.Time)
	keeper.SetPrice(
		ctx, input.addresses[1], input.stdAssetCode,
		sdk.NewInt(35000000),
		header.Time)
	keeper.SetPrice(
		ctx, input.addresses[2], input.stdAssetCode,
		sdk.NewInt(34000000),
		header.Time)
	keeper.SetPrice(
		ctx, input.addresses[3], input.stdAssetCode,
		sdk.NewInt(36000000),
		header.Time)
	err = keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	price = keeper.GetCurrentPrice(ctx, input.stdAssetCode)
	require.Equal(t, price.Price.Equal(sdk.NewInt(34500000)), true)
}
