// +build unit

package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/utils"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func NewMockCurrentPrice(assetCode string, ask, bid int64) types.CurrentPrice {
	return types.CurrentPrice{
		AssetCode:  dnTypes.AssetCode(assetCode),
		AskPrice:   sdk.NewInt(ask),
		BidPrice:   sdk.NewInt(bid),
		ReceivedAt: time.Now(),
	}
}

// Check CheckPriceReceiveTime method with different timestamp sets.
func TestOracleKeeper_CheckPriceReceiveTime(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	receivedAtDiffDur := time.Duration(keeper.GetPostPriceParams(ctx).ReceivedAtDiffInS) * time.Second

	// check equal timestamps
	{
		require.Nil(t, keeper.checkPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time))
	}

	// check timestamps within +-range
	{
		dur := receivedAtDiffDur / 2
		require.Nil(t, keeper.checkPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(dur)))
		require.Nil(t, keeper.checkPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(-dur)))
	}

	// check timestamps outside of +-range
	{
		dur := receivedAtDiffDur + 1*time.Second
		utils.CheckExpectedErr(t, types.ErrInvalidReceivedAt, keeper.checkPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(dur)))
		utils.CheckExpectedErr(t, types.ErrInvalidReceivedAt, keeper.checkPriceReceivedAtTimestamp(ctx, ctx.BlockHeader().Time.Add(-dur)))
	}
}

// Check SetPrice method, checking various sets of arguments.
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
			sdk.NewInt(33000001),
			sdk.NewInt(33000000),
			header.Time)
		require.NoError(t, err)

		// Get raw prices
		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 1)
		require.Equal(t, rawPrices[0].AskPrice.Equal(sdk.NewInt(33000001)), true)
		require.Equal(t, rawPrices[0].BidPrice.Equal(sdk.NewInt(33000000)), true)
	}

	// Set price by oracle 2
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[1], input.stdAssetCode,
			sdk.NewInt(35000005),
			sdk.NewInt(35000000),
			header.Time)
		require.NoError(t, err)

		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 2)
		require.Equal(t, rawPrices[1].AskPrice.Equal(sdk.NewInt(35000005)), true)
		require.Equal(t, rawPrices[1].BidPrice.Equal(sdk.NewInt(35000000)), true)
	}

	// Update Price by oracle 1
	{
		_, err := keeper.SetPrice(
			ctx, input.addresses[0], input.stdAssetCode,
			sdk.NewInt(37000007),
			sdk.NewInt(37000000),
			header.Time)
		require.NoError(t, err)

		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, rawPrices[0].AskPrice.Equal(sdk.NewInt(37000007)), true)
		require.Equal(t, rawPrices[0].BidPrice.Equal(sdk.NewInt(37000000)), true)
	}
}

// Check GetRawPrice method with a valid scenario.
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
			sdk.NewInt(33000033),
			sdk.NewInt(33000000),
			header.Time)
		require.NoError(t, err)

		// Get raw prices
		rawPrices := keeper.GetRawPrices(ctx, input.stdAssetCode, header.Height)
		require.Equal(t, len(rawPrices), 1)
		require.Equal(t, rawPrices[0].AskPrice.Equal(sdk.NewInt(33000033)), true)
		require.Equal(t, rawPrices[0].BidPrice.Equal(sdk.NewInt(33000000)), true)
	}
}

// Check CurrentPrice method and finding average price for different numbers of oracles.
func TestOracleKeeper_CurrentPrice(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx
	header := ctx.BlockHeader()

	_, _ = keeper.SetPrice(
		ctx, input.addresses[0], input.stdAssetCode,
		sdk.NewInt(33000001),
		sdk.NewInt(33000000),
		header.Time)
	_, _ = keeper.SetPrice(
		ctx, input.addresses[1], input.stdAssetCode,
		sdk.NewInt(35000005),
		sdk.NewInt(35000000),
		header.Time)
	_, _ = keeper.SetPrice(
		ctx, input.addresses[2], input.stdAssetCode,
		sdk.NewInt(34000004),
		sdk.NewInt(34000000),
		header.Time)
	// Set current price
	err := keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	// Get Current price
	price := keeper.GetCurrentPrice(ctx, input.stdAssetCode)
	require.Equal(t, price.AskPrice.Equal(sdk.NewInt(34000004)), true)
	require.Equal(t, price.BidPrice.Equal(sdk.NewInt(34000000)), true)

	// Even number of oracles
	_, _ = keeper.SetPrice(
		ctx, input.addresses[0], input.stdAssetCode,
		sdk.NewInt(33000003),
		sdk.NewInt(33000000),
		header.Time)
	_, _ = keeper.SetPrice(
		ctx, input.addresses[1], input.stdAssetCode,
		sdk.NewInt(35000005),
		sdk.NewInt(35000000),
		header.Time)
	_, _ = keeper.SetPrice(
		ctx, input.addresses[2], input.stdAssetCode,
		sdk.NewInt(34000004),
		sdk.NewInt(34000000),
		header.Time)
	_, _ = keeper.SetPrice(
		ctx, input.addresses[3], input.stdAssetCode,
		sdk.NewInt(36000006),
		sdk.NewInt(36000000),
		header.Time)

	// Checking SetCurrentPrices method
	err = keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)

	// Checking GetCurrentPrice method
	price = keeper.GetCurrentPrice(ctx, input.stdAssetCode)
	ask, ok := sdk.NewIntFromString("34500004")
	require.True(t, ok)
	require.Equal(t, price.AskPrice.Equal(ask), true)

	bid, ok := sdk.NewIntFromString("34500000")
	require.True(t, ok)
	require.Equal(t, price.BidPrice.Equal(bid), true)

	price2 := types.CurrentPrice{
		AssetCode:  dnTypes.AssetCode("usdt_xfi"),
		AskPrice:   sdk.NewInt(1000001),
		BidPrice:   sdk.NewInt(1000000),
		ReceivedAt: time.Now().Add(-1 * time.Hour),
	}

	// Checking addCurrentPrice method
	keeper.addCurrentPrice(ctx, price2)

	// Checking GetCurrentPricesList method
	cpList, err := keeper.GetCurrentPricesList(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(cpList))
	require.Equal(t, cpList[0].AssetCode, price.AssetCode)
	require.Equal(t, cpList[0].AskPrice.Add(cpList[1].AskPrice), price.AskPrice.Add(price2.AskPrice))
	require.Equal(t, cpList[0].BidPrice.Add(cpList[1].BidPrice), price.BidPrice.Add(price2.BidPrice))
}
