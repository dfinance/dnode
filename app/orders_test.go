// +build unit

package app

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders"
)

const (
	queryOrdersListPath = "/custom/orders/list"
)

func TestOrders_Ttl(t *testing.T) {
	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(0), uint8(0)
	baseSupply, quoteSupply := sdk.NewInt(1000), sdk.NewInt(1000)

	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genValidators, _, _, _ := CreateGenAccounts(3, GenDefCoins(t))
	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	clientAddr := genValidators[0].Address
	tester := NewOrderBookTester(t, app)

	marketID := dnTypes.ID{}
	// init currencies and clients
	{
		tester.BeginBlock()

		marketID = tester.RegisterMarket(clientAddr, baseDenom, baseDecimals, quoteDenom, quoteDecimals)
		tester.AddClient(clientAddr, baseSupply, quoteSupply)

		tester.EndBlock()

		acc := app.accountKeeper.GetAccount(GetContext(app, true), clientAddr)
		t.Logf("acc %q: %v", clientAddr.String(), acc.GetCoins())
	}

	var longTtlOrderID dnTypes.ID
	// add orders
	{
		tester.BeginBlock()

		tester.AddSellOrder(clientAddr, marketID, sdk.OneUint(), sdk.OneUint(), 1)
		longTtlOrderID = tester.AddSellOrder(clientAddr, marketID, sdk.OneUint(), sdk.OneUint(), 10)

		tester.EndBlock()
	}

	// check orders exist
	{
		request := orders.OrdersReq{Page: 1, Limit: 10}
		response := orders.Orders{}
		CheckRunQuery(t, app, request, queryOrdersListPath, &response)

		require.Len(t, response, 2)
	}

	// emulate TTL and recheck orders existence
	{
		acc := app.accountKeeper.GetAccount(GetContext(app, true), clientAddr)
		t.Logf("acc %q: %v", clientAddr.String(), acc.GetCoins())

		tester.BeginBlockWithDuration(2 * time.Second)
		tester.EndBlock()

		request := orders.OrdersReq{Page: 1, Limit: 10}
		response := orders.Orders{}
		CheckRunQuery(t, app, request, queryOrdersListPath, &response)

		require.Len(t, response, 1)
		require.True(t, response[0].ID.Equal(longTtlOrderID))
	}
}
