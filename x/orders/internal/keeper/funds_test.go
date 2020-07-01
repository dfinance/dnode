// +build unit

package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	helperTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func Test_Keeper_OrderFill(t *testing.T) {
	input := NewTestInput(t)

	// create market
	market, err := input.marketKeeper.Add(input.ctx, input.baseBtcDenom, input.quoteDenom)
	require.NoError(t, err)

	// create account with supplies
	_, _, addr := authTypes.KeyTestPubAddr()
	curBaseBalance, ok := sdk.NewIntFromString("100000000000") // 1000 btc
	require.True(t, ok)
	curQuoteBalance, ok := sdk.NewIntFromString("1000000000000000000000") // 1000 dfi
	require.True(t, ok)

	acc := input.accountKeeper.NewAccountWithAddress(input.ctx, addr)
	err = acc.SetCoins(
		sdk.Coins{
			sdk.Coin{
				Denom:  input.baseBtcDenom,
				Amount: curBaseBalance,
			}, sdk.Coin{
				Denom:  input.quoteDenom,
				Amount: curQuoteBalance,
			},
		},
	)
	require.NoError(t, err)
	input.accountKeeper.SetAccount(input.ctx, acc)

	assetCode := helperTypes.AssetCode(market.GetAssetCode())

	// post orders
	askPrice := sdk.NewUintFromString("10000000000000000000") // 10 dfi
	askQuantity := sdk.NewUintFromString("5000000000")        // 50 btc
	askOrder, err := input.keeper.PostOrder(input.ctx, addr, assetCode, types.Ask, askPrice, askQuantity, 60)
	require.NoError(t, err)

	bidPrice := sdk.NewUintFromString("25000000000000000000") // 25 dfi
	bidQuantity := sdk.NewUintFromString("2500000000")        // 25 btc
	bidOrder, err := input.keeper.PostOrder(input.ctx, addr, assetCode, types.Bid, bidPrice, bidQuantity, 60)
	require.NoError(t, err)

	now := time.Now()
	askOrder.UpdatedAt, bidOrder.UpdatedAt = now, now
	curBaseBalance, curQuoteBalance = input.GetAccountBalance(addr, input.baseBtcDenom)

	// check partial orders fill
	{
		// ask order
		{
			clearancePrice := sdk.NewUintFromString("15000000000000000000") // 15 dfi
			fillQuantity := askQuantity.Quo(sdk.NewUint(2))
			unfillQuantity := askOrder.Quantity.Sub(fillQuantity)
			fill := types.OrderFill{
				Order:            askOrder,
				ClearancePrice:   clearancePrice,
				QuantityFilled:   fillQuantity,
				QuantityUnfilled: unfillQuantity,
			}
			input.keeper.ExecuteOrderFills(input.ctx, types.OrderFills{fill})

			// check order exists and updated
			require.True(t, input.keeper.Has(input.ctx, askOrder.ID))
			updOrder, err := input.keeper.Get(input.ctx, askOrder.ID)
			require.NoError(t, err)

			require.True(t, updOrder.Quantity.Equal(unfillQuantity))
			require.False(t, updOrder.UpdatedAt.Equal(askOrder.UpdatedAt))

			// check account balance
			fillCoin, err := fill.FillCoin()
			require.NoError(t, err)
			fillQuoteQuantity := fillCoin.Amount

			doRefund, refundCoin, err := fill.RefundCoin()
			require.NoError(t, err)
			require.False(t, doRefund)
			require.Nil(t, refundCoin)

			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)
			require.True(t, orderBaseBalance.Equal(curBaseBalance))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance.Add(fillQuoteQuantity)))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}

		// bid order
		{
			clearancePrice := sdk.NewUintFromString("20000000000000000000") // 20 dfi (with refund)
			fillQuantity := bidQuantity.Quo(sdk.NewUint(2))
			unfillQuantity := bidOrder.Quantity.Sub(fillQuantity)
			fill := types.OrderFill{
				Order:            bidOrder,
				ClearancePrice:   clearancePrice,
				QuantityFilled:   fillQuantity,
				QuantityUnfilled: unfillQuantity,
			}
			input.keeper.ExecuteOrderFills(input.ctx, types.OrderFills{fill})

			// check order exists and updated
			require.True(t, input.keeper.Has(input.ctx, bidOrder.ID))
			updOrder, err := input.keeper.Get(input.ctx, bidOrder.ID)
			require.NoError(t, err)

			require.True(t, updOrder.Quantity.Equal(unfillQuantity))
			require.False(t, updOrder.UpdatedAt.Equal(bidOrder.UpdatedAt))

			// check account balance
			fillCoin, err := fill.FillCoin()
			require.NoError(t, err)
			fillBaseQuantity := fillCoin.Amount

			doRefund, refundCoin, err := fill.RefundCoin()
			require.NoError(t, err)
			require.True(t, doRefund)
			require.NotNil(t, refundCoin)
			refundQuoteQuantity := refundCoin.Amount

			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)
			require.True(t, orderBaseBalance.Equal(curBaseBalance.Add(fillBaseQuantity)))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance.Add(refundQuoteQuantity)))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}
	}

	// check full order fill
	{
		// ask order
		{
			clearancePrice := sdk.NewUintFromString("20000000000000000000") // 20 dfi
			fillQuantity := askQuantity.Quo(sdk.NewUint(2))
			fill := types.OrderFill{
				Order:            askOrder,
				ClearancePrice:   clearancePrice,
				QuantityFilled:   fillQuantity,
				QuantityUnfilled: sdk.ZeroUint(),
			}
			input.keeper.ExecuteOrderFills(input.ctx, types.OrderFills{fill})

			// check order doesn't exist
			require.False(t, input.keeper.Has(input.ctx, askOrder.ID))

			// check account balance
			fillCoin, err := fill.FillCoin()
			require.NoError(t, err)
			fillQuoteQuantity := fillCoin.Amount

			doRefund, refundCoin, err := fill.RefundCoin()
			require.NoError(t, err)
			require.False(t, doRefund)
			require.Nil(t, refundCoin)

			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)
			require.True(t, orderBaseBalance.Equal(curBaseBalance))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance.Add(fillQuoteQuantity)))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}

		// bid order
		{
			clearancePrice := sdk.NewUintFromString("25000000000000000000") // 25 dfi (no refund)
			fillQuantity := bidQuantity.Quo(sdk.NewUint(2))
			fill := types.OrderFill{
				Order:            bidOrder,
				ClearancePrice:   clearancePrice,
				QuantityFilled:   fillQuantity,
				QuantityUnfilled: sdk.ZeroUint(),
			}
			input.keeper.ExecuteOrderFills(input.ctx, types.OrderFills{fill})

			// check order doesn't exist
			require.False(t, input.keeper.Has(input.ctx, bidOrder.ID))

			// check account balance
			fillCoin, err := fill.FillCoin()
			require.NoError(t, err)
			fillBaseQuantity := fillCoin.Amount

			doRefund, refundCoin, err := fill.RefundCoin()
			require.NoError(t, err)
			require.False(t, doRefund)
			require.Nil(t, refundCoin)

			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)
			require.True(t, orderBaseBalance.Equal(curBaseBalance.Add(fillBaseQuantity)))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}
	}
}
