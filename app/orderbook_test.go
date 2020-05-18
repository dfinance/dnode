// +build unit

package app

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func Test_OB_BasicNoDecimalAssets(t *testing.T) {
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

	client0Addr, client1Addr := genValidators[0].Address, genValidators[1].Address
	tester := NewOrderBookTester(t, app)

	marketID := dnTypes.ID{}
	// init currencies and clients
	{
		tester.BeginBlock()

		marketID = tester.RegisterMarket(client0Addr, baseDenom, baseDecimals, quoteDenom, quoteDecimals)
		tester.AddClient(client0Addr, baseSupply, quoteSupply)
		tester.AddClient(client1Addr, baseSupply, quoteSupply)

		tester.EndBlock()
	}

	// add orders
	{
		tester.BeginBlock()

		ask1ID := tester.AddSellOrder(client0Addr, marketID, sdk.NewUint(5), sdk.NewUint(100))
		tester.SetOrderFullFillOutput(client0Addr, ask1ID)

		bid1ID := tester.AddBuyOrder(client1Addr, marketID, sdk.NewUint(5), sdk.NewUint(200))
		tester.SetOrderPartialFillOutput(client1Addr, bid1ID, sdk.NewUint(100))

		// client 0:
		//   * ask order should be fully filled:
		//     * -(100.0) base
		//     * +(5.0 * 100.0) quote
		tester.SetClientOutputCoin(client0Addr, baseDenom, sdk.NewInt(900))
		tester.SetClientOutputCoin(client0Addr, quoteDenom, sdk.NewInt(1500))
		// client 1:
		//   * bid order should be partially filled:
		//     * +(100.0) base (only half filled)
		//     * -(5.0 * 100.0) quote (filled)
		//     * -(5.0 * 100.0) quote (still locked)
		tester.SetClientOutputCoin(client1Addr, baseDenom, sdk.NewInt(1100))
		tester.SetClientOutputCoin(client1Addr, quoteDenom, sdk.NewInt(0))

		tester.EndBlock()
	}

	tester.CheckOrdersOutput()
	tester.CheckClientsOutput()
}

func Test_OB_BasicDiffDecimalAssets(t *testing.T) {
	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(1), uint8(3)
	baseSupply, quoteSupply := sdk.NewInt(1000), sdk.NewInt(1000)

	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genValidators, _, _, _ := CreateGenAccounts(3, GenDefCoins(t))
	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	client0Addr, client1Addr := genValidators[0].Address, genValidators[1].Address
	tester := NewOrderBookTester(t, app)

	marketID := dnTypes.ID{}
	// init currencies and clients
	{
		tester.BeginBlock()

		marketID = tester.RegisterMarket(client0Addr, baseDenom, baseDecimals, quoteDenom, quoteDecimals)
		tester.AddClient(client0Addr, baseSupply, quoteSupply)
		tester.AddClient(client1Addr, baseSupply, quoteSupply)

		tester.EndBlock()
	}

	// add orders
	{
		tester.BeginBlock()

		ask1ID := tester.AddSellOrder(client0Addr, marketID, sdk.NewUint(5), sdk.NewUint(100))
		tester.SetOrderFullFillOutput(client0Addr, ask1ID)

		bid1ID := tester.AddBuyOrder(client1Addr, marketID, sdk.NewUint(5), sdk.NewUint(200))
		tester.SetOrderPartialFillOutput(client1Addr, bid1ID, sdk.NewUint(100))

		// client 0:
		//   * ask order should be fully filled:
		//     * -(10.0) base
		//     * +(0.005 * 10.0) quote
		tester.SetClientOutputCoin(client0Addr, baseDenom, sdk.NewInt(900))
		tester.SetClientOutputCoin(client0Addr, quoteDenom, sdk.NewInt(1050))
		// client 1:
		//   * bid order should be partially filled:
		//     * +(10.0) base (only half filled)
		//     * -(0.005 * 10.0) quote (filled)
		//     * -(0.005 * 10.0) quote (still locked)
		tester.SetClientOutputCoin(client1Addr, baseDenom, sdk.NewInt(1100))
		tester.SetClientOutputCoin(client1Addr, quoteDenom, sdk.NewInt(900))

		tester.EndBlock()
	}

	tester.CheckOrdersOutput()
	tester.CheckClientsOutput()
}

func Test_OB_ManyOrders(t *testing.T) {
	const (
		inputOrdersCount = 10000
		inputMinBaseQ    = 1
		inputMaxBaseQ    = 500
		inputMinQuoteP   = 1
		inputMaxQuoteP   = 1000
	)

	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(0), uint8(0)

	baseSupply, quoteSupply := sdk.ZeroInt(), sdk.ZeroInt()
	var askPrices, askQuantities, bidPrices, bidQuantities []sdk.Uint
	// calculate initial conditions
	//   * baseAsset, quoteAsset supplies
	//   * ask / bid orders prices / quantities distributions
	{
		askPDistr, askQDistr, err := LinerDistribution(ascDistribution, inputOrdersCount, inputMinQuoteP, inputMaxQuoteP, inputMinBaseQ, inputMaxBaseQ)
		require.NoError(t, err, "ask orders distribution build")

		bidPDistr, bidQDistr, err := LinerDistribution(descDistribution, inputOrdersCount, inputMinQuoteP, inputMaxQuoteP, inputMinBaseQ, inputMaxBaseQ)
		require.NoError(t, err, "bid orders distribution build")

		askPs, askMaxP := Float64ToUintSlice(askPDistr)
		askQs, askMaxQ := Float64ToUintSlice(askQDistr)
		bidPs, bidMaxP := Float64ToUintSlice(bidPDistr)
		bidQs, bidMaxQ := Float64ToUintSlice(bidQDistr)

		askPrices, askQuantities = askPs, askQs
		bidPrices, bidQuantities = bidPs, bidQs

		baseSupply = sdk.Int(bidMaxQ)
		if askMaxQ.GT(bidMaxQ) {
			baseSupply = sdk.Int(askMaxQ)
		}

		quoteSupply = sdk.Int(bidMaxP)
		if askMaxP.GT(bidMaxP) {
			quoteSupply = sdk.Int(askMaxP)
		}
		quoteSupply = quoteSupply.Mul(baseSupply)
	}
	t.Logf("BaseAsset supply: %s", baseSupply.String())
	t.Logf("QuoteAsset supply: %s", quoteSupply.String())

	t.Parallel()
	app, server := newTestDnApp(log.AllowError())
	defer app.CloseConnections()
	defer server.Stop()

	genValidators, _, _, _ := CreateGenAccounts(3, GenDefCoins(t))
	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	client0Addr, client1Addr := genValidators[0].Address, genValidators[1].Address
	tester := NewOrderBookTester(t, app)

	marketID := dnTypes.ID{}
	// init currencies and clients
	{
		tester.BeginBlock()

		marketID = tester.RegisterMarket(client0Addr, baseDenom, baseDecimals, quoteDenom, quoteDecimals)
		tester.AddClient(client0Addr, baseSupply, quoteSupply)
		tester.AddClient(client1Addr, baseSupply, quoteSupply)

		tester.EndBlock()
	}

	// add and process orders
	processingDur := time.Duration(0)
	{
		tester.BeginBlock()

		for i := uint(0); i < inputOrdersCount; i++ {
			tester.AddSellOrder(client0Addr, marketID, askPrices[i], askQuantities[i])
		}
		t.Logf("Ask orders [first]: P -> Q: %s -> %s", askPrices[0], askQuantities[0])
		t.Logf("Ask orders [last]: P -> Q: %s -> %s", askPrices[len(askPrices)-1], askQuantities[len(askQuantities)-1])

		for i := uint(0); i < inputOrdersCount; i++ {
			tester.AddBuyOrder(client1Addr, marketID, bidPrices[i], bidQuantities[i])
		}
		t.Logf("Bid orders [first]: P -> Q: %s -> %s", bidPrices[0], bidQuantities[0])
		t.Logf("Bid orders [last]: P -> Q: %s -> %s", bidPrices[len(bidPrices)-1], bidQuantities[len(bidQuantities)-1])

		processingStart := time.Now()
		tester.EndBlock()
		processingDur = time.Now().Sub(processingStart)
	}

	tester.CheckOrdersOutput()
	tester.CheckClientsOutput()
	tester.PrintHistoryItems()

	t.Logf("Orders %d -> %v", inputOrdersCount, processingDur)
}
