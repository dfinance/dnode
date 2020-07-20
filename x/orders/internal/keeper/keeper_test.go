// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/perms"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

func TestOrdersKeeper_PostRevokeOrder(t *testing.T) {
	input := NewTestInput(
		t,
		perms.Permissions{
			markets.PermCreator,
			markets.PermReader,
		},
	)

	// non-existing market
	{
		owner := sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07z`")
		_, err := input.keeper.PostOrder(input.ctx, owner, dnTypes.AssetCode("dfi_usd"), types.Bid, sdk.OneUint(), sdk.OneUint(), 60)
		require.Error(t, err)
	}

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

	// post orders
	var askOrder, bidOrder types.Order
	{
		// ask order
		{
			price := sdk.OneUint()
			quantity := sdk.NewUintFromString("1000000000") // 10 btc

			// post and check returned order
			postOrder, err := input.keeper.PostOrder(input.ctx, addr, market.GetAssetCode(), types.Ask, price, quantity, 60)
			require.NoError(t, err)
			require.Equal(t, postOrder.ID.UInt64(), uint64(0))
			require.True(t, postOrder.Market.ID.Equal(market.ID))
			require.True(t, postOrder.Direction.Equal(types.Ask))
			require.True(t, postOrder.Price.Equal(price))
			require.True(t, postOrder.Quantity.Equal(quantity))

			// compare returned and stored orders
			readOrder, err := input.keeper.Get(input.ctx, postOrder.ID)
			require.NoError(t, err)
			CompareOrders(t, postOrder, readOrder)

			// check owner balance (funds lock)
			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)

			lockCoin, err := postOrder.LockCoin()
			require.NoError(t, err)
			baseLockQuantity := lockCoin.Amount

			require.True(t, orderQuoteBalance.Equal(curQuoteBalance))
			require.True(t, orderBaseBalance.Equal(curBaseBalance.Sub(baseLockQuantity)))

			askOrder = postOrder
			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}

		// bid order
		{
			price := sdk.NewUintFromString("10000000000000000000") // 10 dfi
			quantity := sdk.NewUintFromString("1000000000")        // 10 btc

			// post and check returned order
			postOrder, err := input.keeper.PostOrder(input.ctx, addr, market.GetAssetCode(), types.Bid, price, quantity, 60)
			require.NoError(t, err)
			require.Equal(t, postOrder.ID.UInt64(), uint64(1))
			require.True(t, postOrder.Market.ID.Equal(market.ID))
			require.True(t, postOrder.Direction.Equal(types.Bid))
			require.True(t, postOrder.Price.Equal(price))
			require.True(t, postOrder.Quantity.Equal(quantity))

			// compare returned and stored orders
			readOrder, err := input.keeper.Get(input.ctx, postOrder.ID)
			require.NoError(t, err)
			CompareOrders(t, postOrder, readOrder)

			// check owner balance (funds lock)
			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)

			lockCoin, err := postOrder.LockCoin()
			require.NoError(t, err)
			quoteLockQuantity := lockCoin.Amount

			require.True(t, orderBaseBalance.Equal(curBaseBalance))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance.Sub(quoteLockQuantity)))

			bidOrder = postOrder
			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}
	}

	// revoke non-existing order
	{
		err := input.keeper.RevokeOrder(input.ctx, dnTypes.NewIDFromUint64(2))
		require.Error(t, err)
	}

	// cancel non-existing order
	{
		err := input.keeper.RevokeOrder(input.ctx, dnTypes.NewIDFromUint64(2))
		require.Error(t, err)
	}

	// cancel existing order
	{
		// ask order
		{
			// revoke
			err := input.keeper.RevokeOrder(input.ctx, askOrder.ID)
			require.NoError(t, err)

			// check order removed
			require.False(t, input.keeper.Has(input.ctx, askOrder.ID))
			require.True(t, input.keeper.Has(input.ctx, bidOrder.ID))

			orders, err := input.keeper.GetList(input.ctx)
			require.NoError(t, err)
			require.Len(t, orders, 1)
			CompareOrders(t, bidOrder, orders[0])

			// check owner balance (funds unlock)
			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)

			unlockCoin, err := askOrder.LockCoin()
			require.NoError(t, err)
			baseUnlockQuantity := unlockCoin.Amount

			require.True(t, orderBaseBalance.Equal(curBaseBalance.Add(baseUnlockQuantity)))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}

		// bid order
		{
			// revoke
			err := input.keeper.RevokeOrder(input.ctx, bidOrder.ID)
			require.NoError(t, err)

			// check order removed
			require.False(t, input.keeper.Has(input.ctx, bidOrder.ID))

			orders, err := input.keeper.GetList(input.ctx)
			require.NoError(t, err)
			require.Len(t, orders, 0)

			// check owner balance (funds unlock)
			orderBaseBalance, orderQuoteBalance := input.GetAccountBalance(addr, input.baseBtcDenom)

			unlockCoin, err := bidOrder.LockCoin()
			require.NoError(t, err)
			quoteUnlockQuantity := unlockCoin.Amount

			require.True(t, orderBaseBalance.Equal(curBaseBalance))
			require.True(t, orderQuoteBalance.Equal(curQuoteBalance.Add(quoteUnlockQuantity)))

			curBaseBalance, curQuoteBalance = orderBaseBalance, orderQuoteBalance
		}
	}
}
