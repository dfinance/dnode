package app

import (
	"encoding/hex"
	"fmt"
	"math"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register"
	marketTypes "github.com/dfinance/dnode/x/markets"
	obTypes "github.com/dfinance/dnode/x/orderbook"
	orderTypes "github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/vmauth"
)

const (
	queryMarketsListPath = "/custom/markets/list"
	queryOrdersListPath  = "/custom/orders/list"
	//
	ascDistribution  = 1
	descDistribution = 2
)

type OrderBookTester struct {
	t   *testing.T
	app *DnServiceApp
	// clients slice
	Clients []*ClientTestState
	// markets maps (key: ID)
	Markets map[string]marketTypes.Market
	// currencies info map (key: denom)
	Currencies map[string]currencies_register.CurrencyInfo
}

type ClientTestState struct {
	Address sdk.AccAddress
	// initial coin balances
	InputCoins map[string]sdk.Int
	// output check coin balances
	OutputCoins map[string]sdk.Int
	// coin balances calculated based on filled orders (initial value equal to InputCoins)
	EstimatedCoins map[string]sdk.Int
	// client orders slice
	Orders []*OrderTestState
}

type OrderTestState struct {
	ID dnTypes.ID
	// order post state
	Input OrderInput
	// output check state
	Output OrderOutput
}

type OrderInput struct {
	MarketID  dnTypes.ID
	Direction orderTypes.Direction
	Price     sdk.Uint
	Quantity  sdk.Uint
	CreatedAt time.Time
}

type OrderOutput struct {
	// check order output flag
	Check bool
	// order should be fully filled flag
	FullyFilled bool
	// quantity order should be filled to if partially filled
	PartialQuantity sdk.Uint
}

func NewOrderBookTester(t *testing.T, app *DnServiceApp) OrderBookTester {
	tester := OrderBookTester{
		t:          t,
		app:        app,
		Markets:    make(map[string]marketTypes.Market, 0),
		Currencies: make(map[string]currencies_register.CurrencyInfo, 0),
		Clients:    make([]*ClientTestState, 0),
	}

	return tester
}

// Start a new block.
func (tester *OrderBookTester) BeginBlock() {
	tester.app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			ChainID: chainID,
			Height:  tester.app.LastBlockHeight() + 1,
		},
	})
}

// End current block.
func (tester *OrderBookTester) EndBlock() {
	tester.app.EndBlock(abci.RequestEndBlock{})
	tester.app.Commit()

	return
}

// Add new currencies and register a corresponding market.
func (tester *OrderBookTester) RegisterMarket(ownerAddr sdk.AccAddress, baseDenom string, baseDecimals uint8, quoteDenom string, quoteDecimals uint8) (marketID dnTypes.ID) {
	ctx := GetContext(tester.app, false)

	registerCurrency := func(denom string, decimals uint8) {
		path := hex.EncodeToString([]byte(denom))

		require.NoError(tester.t, vmauth.AddDenomPath(denom, string(path)), "registering path for denom: %s", denom)
		err := tester.app.crKeeper.AddCurrencyInfo(
			ctx,
			denom,
			decimals,
			false,
			common_vm.Bech32ToLibra(ownerAddr),
			sdk.ZeroInt(),
			[]byte(denom),
		)
		require.NoError(tester.t, err, "adding currency for denom: %s", denom)

		ccInfo, err := tester.app.crKeeper.GetCurrencyInfo(ctx, denom)
		require.NoError(tester.t, err, "checking currency added for denom: %s", denom)
		tester.Currencies[denom] = ccInfo
	}

	// register currencies
	{
		registerCurrency(baseDenom, baseDecimals)
		registerCurrency(quoteDenom, quoteDecimals)
	}

	// register market
	{
		market, err := tester.app.marketKeeper.Add(ctx, baseDenom, quoteDenom)
		require.NoError(tester.t, err, "registering market for assets: %s / %s", baseDenom, quoteDenom)

		marketID = market.ID
		tester.Markets[marketID.String()] = market
	}

	// check market created
	{
		marketExt, err := tester.app.marketKeeper.GetExtended(ctx, marketID)
		require.NoError(tester.t, err, "getting marketExt for assets: %s / %s", baseDenom, quoteDenom)
		require.Equal(tester.t, baseDenom, string(marketExt.BaseCurrency.Denom))
		require.Equal(tester.t, baseDecimals, marketExt.BaseCurrency.Decimals)
		require.Equal(tester.t, quoteDenom, string(marketExt.QuoteCurrency.Denom))
		require.Equal(tester.t, quoteDecimals, marketExt.QuoteCurrency.Decimals)
	}

	return
}

// Add a client to Tester, set initial coin balances.
func (tester *OrderBookTester) AddClient(addr sdk.AccAddress, baseCoinsAmount, quoteCoinsAmount sdk.Int) {
	ctx := GetContext(tester.app, false)
	clientID := len(tester.Clients)

	// append all markets assets coins to account
	{
		acc := tester.app.accountKeeper.GetAccount(ctx, addr)
		accCoins := acc.GetCoins()
		for _, market := range tester.Markets {
			accCoins = append(accCoins, sdk.Coin{
				Denom:  market.BaseAssetDenom,
				Amount: baseCoinsAmount,
			})
			accCoins = append(accCoins, sdk.Coin{
				Denom:  market.QuoteAssetDenom,
				Amount: quoteCoinsAmount,
			})
		}
		require.NoError(tester.t, acc.SetCoins(accCoins), "setting coins for client: %d", clientID)

		tester.app.accountKeeper.SetAccount(ctx, acc)
	}

	// save coins client input
	{
		acc := tester.app.accountKeeper.GetAccount(ctx, addr)
		clientState := &ClientTestState{
			Address:        addr,
			InputCoins:     make(map[string]sdk.Int, 0),
			OutputCoins:    make(map[string]sdk.Int, 0),
			EstimatedCoins: make(map[string]sdk.Int, 0),
			Orders:         make([]*OrderTestState, 0),
		}

		for _, coin := range acc.GetCoins() {
			if _, ok := tester.Currencies[coin.Denom]; ok {
				clientState.InputCoins[coin.Denom] = coin.Amount
				clientState.EstimatedCoins[coin.Denom] = coin.Amount
			}
		}
		tester.Clients = append(tester.Clients, clientState)
	}
}

// Set expected output coin balances for client.
func (tester *OrderBookTester) SetClientOutputCoin(clientAddr sdk.AccAddress, denom string, amount sdk.Int) {
	clientState := tester.findClient(clientAddr)
	clientState.OutputCoins[denom] = amount
}

// Add a bid order for client.
func (tester *OrderBookTester) AddBuyOrder(clientAddr sdk.AccAddress, marketID dnTypes.ID, price, quantity sdk.Uint) (orderID dnTypes.ID) {
	orderState := tester.addOrder(clientAddr, orderTypes.BidDirection, marketID, price, quantity)
	return orderState.ID
}

// Add an ask order for client.
func (tester *OrderBookTester) AddSellOrder(clientAddr sdk.AccAddress, marketID dnTypes.ID, price, quantity sdk.Uint) (orderID dnTypes.ID) {
	orderState := tester.addOrder(clientAddr, orderTypes.AskDirection, marketID, price, quantity)
	return orderState.ID
}

// Set expected order output (fully filled).
func (tester *OrderBookTester) SetOrderFullFillOutput(clientAddr sdk.AccAddress, id dnTypes.ID) {
	orderState := tester.findOrder(clientAddr, id)
	orderState.Output.Check = true
	orderState.Output.FullyFilled = true
}

// Set expected order output (partially filled).
func (tester *OrderBookTester) SetOrderPartialFillOutput(clientAddr sdk.AccAddress, id dnTypes.ID, quantity sdk.Uint) {
	orderState := tester.findOrder(clientAddr, id)
	orderState.Output.Check = true
	orderState.Output.PartialQuantity = quantity
}

// Get current clients balances and compare to expected output (if provided).
func (tester *OrderBookTester) CheckClientsOutput() {
	tester.t.Log()

	// iterate over all clients
	for clientID, clientState := range tester.Clients {
		tester.t.Logf("Client %d (%s): coin results:", clientID, clientState.Address)

		// get current balances
		acc := GetAccountCheckTx(tester.app, clientState.Address)
		outputCoins := acc.GetCoins()

		// iterate over all client coin initial balances
		for denom, inputAmount := range clientState.InputCoins {
			estimatedAmount := clientState.EstimatedCoins[denom]

			// get currency info and print initial balance
			ccInfo := tester.Currencies[denom]
			tester.t.Logf("  %q asset:", denom)
			tester.t.Logf("    initial:      %s (%s)", inputAmount, ccInfo.UintToDec(sdk.Uint(inputAmount)))
			tester.t.Logf("    -> estimated: %s (%s)", estimatedAmount, ccInfo.UintToDec(sdk.Uint(estimatedAmount)))

			// find matching output coin (empty if not found)
			var outputCoin *sdk.Coin
			outputAmount := sdk.ZeroInt()
			for i := 0; i < len(outputCoins); i++ {
				if outputCoins[i].Denom == denom {
					outputCoin = &outputCoins[i]
					break
				}
			}
			if outputCoin != nil {
				tester.t.Logf("    -> actual:    %s (%s)", outputCoin.Amount, ccInfo.UintToDec(sdk.Uint(outputCoin.Amount)))
				outputAmount = outputCoin.Amount
			} else {
				tester.t.Logf("    -> actual:    empty")
			}

			// check estimated balances (after all the orders were processed)
			require.True(tester.t, estimatedAmount.GTE(sdk.ZeroInt()), "estimated amount for client %d is LT 0: %s", clientID, denom)
			require.True(tester.t, outputAmount.Equal(estimatedAmount), "checking coins amount for client %d (estimated): %s", clientID, denom)

			// check output amount if provided
			checkAmount, ok := clientState.OutputCoins[denom]
			if !ok {
				continue
			}
			require.True(tester.t, outputAmount.Equal(checkAmount), "checking coins amount for client %d (output): %s", clientID, denom)
		}
	}
}

// Get current order states and compare to expected output (if provided).
func (tester *OrderBookTester) CheckOrdersOutput() {
	tester.t.Log()

	ctx := GetContext(tester.app, true)
	orders, err := tester.app.orderKeeper.GetList(ctx)
	require.NoError(tester.t, err, "getting orders list")

	// build orders map (to optimize search)
	ordersMap := make(map[string]*orderTypes.Order, len(orders))
	for i := 0; i < len(orders); i++ {
		ordersMap[orders[i].ID.String()] = &orders[i]
	}

	// iterate over all clients
	for clientID, clientState := range tester.Clients {
		tester.t.Logf("Client %d (%s): order results", clientID, clientState.Address)
		// iterate over all client orders
		for _, orderState := range clientState.Orders {
			// prepare inputs for order checking
			marketExt, err := tester.app.marketKeeper.GetExtended(ctx, orderState.Input.MarketID)
			require.NoError(tester.t, err, "getting marketExt: %s", orderState.Input.MarketID)

			historyItem, err := tester.app.orderBookKeeper.GetHistoryItem(ctx, marketExt.ID, ctx.BlockHeight()-1)
			require.NoError(tester.t, err, "getting historyItem for block: %d", ctx.BlockHeight()-1)

			outputOrder := ordersMap[orderState.ID.String()]

			tester.checkOrder(clientState, marketExt, orderState, historyItem, outputOrder)
		}
	}
}

// Print history for the latests block height and all markets.
func (tester *OrderBookTester) PrintHistoryItems() {
	ctx := GetContext(tester.app, true)

	historyItems := obTypes.HistoryItems{}
	for _, m := range tester.Markets {
		historyItem, err := tester.app.orderBookKeeper.GetHistoryItem(ctx, m.ID, ctx.BlockHeight()-1)
		require.NoError(tester.t, err, "getting historyItem for block: %d", ctx.BlockHeight()-1)
		historyItems = append(historyItems, historyItem)
	}

	tester.t.Logf("\n%s", historyItems.String())
}

func (tester *OrderBookTester) checkOrder(clientSt *ClientTestState, market marketTypes.MarketExtended, orderSt *OrderTestState, historyItem obTypes.HistoryItem, orderOut *orderTypes.Order) {
	// print logs
	tester.t.Logf("  Order %s (%s):", orderSt.ID, orderSt.Input.Direction)
	tester.t.Logf("    quantity: %s (%s)", orderSt.Input.Quantity, market.BaseCurrency.UintToDec(orderSt.Input.Quantity))

	if orderOut == nil {
		tester.t.Logf("    -> fully filled")
	} else {
		tester.t.Logf("    -> partially filled: %s (%s)", orderOut.Quantity, market.BaseCurrency.UintToDec(orderOut.Quantity))
	}

	// updated estimated client coin balance based on current order output
	{
		baseDenom := string(market.BaseCurrency.Denom)
		quoteDenom := string(market.QuoteCurrency.Denom)

		// process locked coins
		if orderSt.Input.Direction == orderTypes.BidDirection {
			quoteQuantity, _ := market.BaseToQuoteQuantity(orderSt.Input.Price, orderSt.Input.Quantity)
			quoteCoin := clientSt.EstimatedCoins[quoteDenom]
			clientSt.EstimatedCoins[quoteDenom] = quoteCoin.Sub(sdk.Int(quoteQuantity))
		} else {
			baseCoin := clientSt.EstimatedCoins[baseDenom]
			clientSt.EstimatedCoins[baseDenom] = baseCoin.Sub(sdk.Int(orderSt.Input.Quantity))
		}

		// process baseAsset filled quantity
		baseCoin := clientSt.EstimatedCoins[baseDenom]
		baseFilledQuantity := orderSt.Input.Quantity
		if orderOut != nil {
			baseFilledQuantity = baseFilledQuantity.Sub(orderOut.Quantity)
		}
		if !baseFilledQuantity.IsZero() {
			// if order was filled
			if orderSt.Input.Direction == orderTypes.BidDirection {
				// client should get baseAssets
				clientSt.EstimatedCoins[baseDenom] = baseCoin.Add(sdk.Int(baseFilledQuantity))
			}
		}

		// process baseAsset filled quantity
		quoteCoin := clientSt.EstimatedCoins[quoteDenom]
		if !baseFilledQuantity.IsZero() {
			// if order was filled
			quoteFilledQuantity, _ := market.BaseToQuoteQuantity(historyItem.ClearancePrice, baseFilledQuantity)
			if orderSt.Input.Direction == orderTypes.BidDirection {
				if historyItem.ClearancePrice.LT(orderSt.Input.Price) {
					// client should get refund
					priceDiff := orderSt.Input.Price.Sub(historyItem.ClearancePrice)
					if refundQuoteQuantity, err := market.BaseToQuoteQuantity(priceDiff, baseFilledQuantity); err == nil {
						// refund quantity is big enough
						clientSt.EstimatedCoins[quoteDenom] = quoteCoin.Add(sdk.Int(refundQuoteQuantity))
					}
				}
			} else {
				// client should get quoteAssets
				clientSt.EstimatedCoins[quoteDenom] = quoteCoin.Add(sdk.Int(quoteFilledQuantity))
			}
		}
	}

	// do checks if output is provided
	if !orderSt.Output.Check {
		return
	}

	if orderSt.Output.FullyFilled {
		require.Nil(tester.t, orderOut, "order %s: not fully filled", orderSt.ID)
	} else {
		require.NotNil(tester.t, orderOut, "order %s: not exists", orderSt.ID)
		require.True(tester.t, orderSt.Output.PartialQuantity.Equal(orderOut.Quantity), "order %s: not partially filled", orderSt.ID)
	}
}

func (tester *OrderBookTester) addOrder(owner sdk.AccAddress, dir orderTypes.Direction, mID dnTypes.ID, p, q sdk.Uint) (orderState *OrderTestState) {
	ctx := GetContext(tester.app, false)

	clientState := tester.findClient(owner)
	_, ok := tester.Markets[mID.String()]
	require.True(tester.t, ok, "market not found: %s", mID)

	// post order
	orderID := dnTypes.ID{}
	{
		order, err := tester.app.orderKeeper.PostOrder(
			ctx,
			owner,
			mID,
			dir,
			p,
			q,
			60,
		)
		require.NoError(tester.t, err, "posting order")
		orderID = order.ID
	}

	// check order added
	{
		order, err := tester.app.orderKeeper.Get(ctx, orderID)
		require.NoError(tester.t, err, "getting order")
		require.True(tester.t, owner.Equals(order.Owner))
		require.True(tester.t, order.Market.ID.Equal(mID))
		require.Equal(tester.t, dir.String(), order.Direction.String())
		require.True(tester.t, p.Equal(order.Price))
		require.True(tester.t, q.Equal(order.Quantity))
		require.Equal(tester.t, time.Duration(60)*time.Second, order.Ttl)
		require.True(tester.t, order.CreatedAt.Equal(order.UpdatedAt))

		orderState = &OrderTestState{
			ID: orderID,
			Input: OrderInput{
				MarketID:  mID,
				Direction: dir,
				Price:     p,
				Quantity:  q,
				CreatedAt: order.CreatedAt,
			},
			Output: OrderOutput{},
		}

		clientState.Orders = append(clientState.Orders, orderState)
	}

	return
}

func (tester *OrderBookTester) findClient(clientAddr sdk.AccAddress) (retState *ClientTestState) {
	for i := 0; i < len(tester.Clients); i++ {
		if tester.Clients[i].Address.Equals(clientAddr) {
			retState = tester.Clients[i]
			break
		}
	}
	require.NotNil(tester.t, retState, "client not found: %s", clientAddr)

	return
}

func (tester *OrderBookTester) findOrder(clientAddr sdk.AccAddress, id dnTypes.ID) (retState *OrderTestState) {
	clientState := tester.findClient(clientAddr)
	for i := 0; i < len(clientState.Orders); i++ {
		if clientState.Orders[i].ID.Equal(id) {
			retState = clientState.Orders[i]
			break
		}
	}
	require.NotNil(tester.t, retState, "order not found: %s", id)

	return
}

// Get linear distribution (ASC/DESC) for {count} points with min/max limits.
func LinerDistribution(direction int, count uint, xMin, xMax, yMin, yMax float64) (x, y []float64, retErr error) {
	var yFunc func(xNorm float64) float64
	switch direction {
	case ascDistribution:
		yFunc = func(xNorm float64) float64 {
			return xNorm
		}
	case descDistribution:
		yFunc = func(xNorm float64) float64 {
			return 1.0 - xNorm
		}
	default:
		retErr = fmt.Errorf("unsupported direction")
		return
	}

	return getDistribution(count, xMin, xMax, yMin, yMax, yFunc)
}

// Get x^3 distribution (ASC/DESC) for {count} points with min/max limits.
func Cubic3Distribution(direction int, count uint, xMin, xMax, yMin, yMax float64) (x, y []float64, retErr error) {
	var yFunc func(xNorm float64) float64
	switch direction {
	case ascDistribution:
		yFunc = func(xNorm float64) float64 {
			return math.Pow(xNorm, 3.0)
		}
	case descDistribution:
		yFunc = func(xNorm float64) float64 {
			return -1.0 * math.Pow(xNorm-1.0, 3.0)
		}
	default:
		retErr = fmt.Errorf("unsupported direction")
		return
	}

	return getDistribution(count, xMin, xMax, yMin, yMax, yFunc)
}

// Convert []float64 to []sdk.Uint and calculate sum.
func Float64ToUintSlice(in []float64) (out []sdk.Uint, sum sdk.Uint) {
	sum = sdk.ZeroUint()
	out = make([]sdk.Uint, 0, len(in))

	for _, floatValue := range in {
		uintValue := sdk.NewUint(uint64(math.Round(floatValue)))

		out = append(out, uintValue)
		sum = sum.Add(uintValue)
	}

	return
}

func getDistribution(count uint, xMin, xMax, yMin, yMax float64, yFunc func(xNorm float64) float64) (x, y []float64, retErr error) {
	if xMax <= xMin {
		retErr = fmt.Errorf("xMax LTE xMin")
		return
	}
	if yMax <= yMin {
		retErr = fmt.Errorf("yMax LTE yMin")
		return
	}
	if count == 0 {
		retErr = fmt.Errorf("count EQ 0")
		return
	}

	if yZero := yFunc(0.0); yZero != 0.0 && yZero != 1.0 {
		retErr = fmt.Errorf("invalid yFunc: f(0) NE 0.0 / 1.0")
		return
	}
	if yZero := yFunc(1.0); yZero != 0.0 && yZero != 1.0 {
		retErr = fmt.Errorf("invalid yFunc: f(1) NE 0.0 / 1.0")
		return
	}

	x = make([]float64, 0, count)
	y = make([]float64, 0, count)

	xDiff := xMax - xMin
	yDiff := yMax - yMin
	stepX := (xMax - xMin) / float64(count-1)
	for i := uint(0); i < count; i++ {
		xCur := xMin + float64(i)*stepX
		xNorm := (xCur - xMin) / xDiff

		yCur := yMin + yDiff*yFunc(xNorm)

		x = append(x, xCur)
		y = append(y, yCur)
	}

	return
}
