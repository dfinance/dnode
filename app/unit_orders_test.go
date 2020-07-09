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
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genValidators, _, _, _ := CreateGenAccounts(3, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genValidators)

	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(0), uint8(0)
	baseSupply, quoteSupply := sdk.NewInt(1000), sdk.NewInt(1000)

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
		request := orders.OrdersReq{Page: sdk.NewUint(1), Limit: sdk.NewUint(10)}
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

		request := orders.OrdersReq{Page: sdk.NewUint(1), Limit: sdk.NewUint(10)}
		response := orders.Orders{}
		CheckRunQuery(t, app, request, queryOrdersListPath, &response)

		require.Len(t, response, 1)
		require.True(t, response[0].ID.Equal(longTtlOrderID))
	}
}
