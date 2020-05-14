// +build unit

package app

import (
	"encoding/hex"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register"
	marketTypes "github.com/dfinance/dnode/x/markets"
	orderTypes "github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/vmauth"
)

const (
	queryMarketsListPath = "/custom/markets/list"
	queryOrdersListPath  = "/custom/orders/list"
)

type OrderBookTester struct {
	t          *testing.T
	app        *DnServiceApp
	Clients    []*ClientTestState
	Markets    map[string]marketTypes.Market
	Currencies map[string]currencies_register.CurrencyInfo
}

type ClientTestState struct {
	Address     sdk.AccAddress
	InputCoins  map[string]sdk.Int
	OutputCoins map[string]sdk.Int
	Orders      []*OrderTestState
}

type OrderTestState struct {
	ID     dnTypes.ID
	Input  OrderInput
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
	Check           bool
	FullyFilled     bool
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

func (tester *OrderBookTester) BeginBlock() {
	tester.app.BeginBlock(abci.RequestBeginBlock{
		Header: abci.Header{
			ChainID: chainID,
			Height:  tester.app.LastBlockHeight() + 1,
		},
	})
}

func (tester *OrderBookTester) EndBlock() {
	tester.app.EndBlock(abci.RequestEndBlock{})
	tester.app.Commit()
}

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

func (tester *OrderBookTester) AddClient(addr sdk.AccAddress, coinsAmount sdk.Int) {
	ctx := GetContext(tester.app, false)
	clientID := len(tester.Clients)

	// append all markets assets coins to account
	{
		acc := tester.app.accountKeeper.GetAccount(ctx, addr)
		accCoins := acc.GetCoins()
		for _, market := range tester.Markets {
			accCoins = append(accCoins, sdk.Coin{
				Denom:  string(market.BaseAssetDenom),
				Amount: coinsAmount,
			})
			accCoins = append(accCoins, sdk.Coin{
				Denom:  string(market.QuoteAssetDenom),
				Amount: coinsAmount,
			})
		}
		require.NoError(tester.t, acc.SetCoins(accCoins), "setting coins for client: %d", clientID)

		tester.app.accountKeeper.SetAccount(ctx, acc)
	}

	// save coins client input
	{
		acc := tester.app.accountKeeper.GetAccount(ctx, addr)
		clientState := &ClientTestState{
			Address:     addr,
			InputCoins:  make(map[string]sdk.Int, 0),
			OutputCoins: make(map[string]sdk.Int, 0),
			Orders:      make([]*OrderTestState, 0),
		}

		for _, coin := range acc.GetCoins() {
			if _, ok := tester.Currencies[coin.Denom]; ok {
				clientState.InputCoins[coin.Denom] = coin.Amount
			}
		}
		tester.Clients = append(tester.Clients, clientState)
	}
}

func (tester *OrderBookTester) SetClientOutputCoin(clientAddr sdk.AccAddress, denom string, amount sdk.Int) {
	clientState := tester.findClient(clientAddr)
	clientState.OutputCoins[denom] = amount
}

func (tester *OrderBookTester) AddBuyOrder(clientAddr sdk.AccAddress, marketID dnTypes.ID, price, quantity sdk.Uint) (orderID dnTypes.ID) {
	orderState := tester.addOrder(clientAddr, orderTypes.BidDirection, marketID, price, quantity)
	return orderState.ID
}

func (tester *OrderBookTester) AddSellOrder(clientAddr sdk.AccAddress, marketID dnTypes.ID, price, quantity sdk.Uint) (orderID dnTypes.ID) {
	orderState := tester.addOrder(clientAddr, orderTypes.AskDirection, marketID, price, quantity)
	return orderState.ID
}

func (tester *OrderBookTester) SetOrderFullFillOutput(clientAddr sdk.AccAddress, id dnTypes.ID) {
	orderState := tester.findOrder(clientAddr, id)
	orderState.Output.Check = true
	orderState.Output.FullyFilled = true
}

func (tester *OrderBookTester) SetOrderPartialFillOutput(clientAddr sdk.AccAddress, id dnTypes.ID, quantity sdk.Uint) {
	orderState := tester.findOrder(clientAddr, id)
	orderState.Output.Check = true
	orderState.Output.PartialQuantity = quantity
}

func (tester *OrderBookTester) CheckClientsOutput() {
	for clientID, clientState := range tester.Clients {
		acc := GetAccountCheckTx(tester.app, clientState.Address)
		outputCoins := acc.GetCoins()

		tester.t.Logf("Client %d (%s): coin results:", clientID, clientState.Address)
		for denom, inputAmount := range clientState.InputCoins {
			ccInfo := tester.Currencies[denom]
			tester.t.Logf("  %q amount: %s (%s)", denom, inputAmount, ccInfo.UintToDec(sdk.Uint(inputAmount)))

			var outputCoin *sdk.Coin
			outputAmount := sdk.ZeroInt()
			for i := 0; i < len(outputCoins); i++ {
				if outputCoins[i].Denom == denom {
					outputCoin = &outputCoins[i]
					break
				}
			}
			if outputCoin != nil {
				tester.t.Logf("  -> %s (%s)", outputCoin.Amount, ccInfo.UintToDec(sdk.Uint(outputCoin.Amount)))
				outputAmount = outputCoin.Amount
			} else {
				tester.t.Logf("  -> empty")
			}

			checkAmount, ok := clientState.OutputCoins[denom]
			if !ok {
				continue
			}
			require.True(tester.t, outputAmount.Equal(checkAmount), "checking coins amount for client %d: %s", clientID, denom)
		}
	}
}

func (tester *OrderBookTester) CheckOrdersOutput() {
	ctx := GetContext(tester.app, true)
	orders, err := tester.app.orderKeeper.List(ctx)
	require.NoError(tester.t, err, "getting orders list")

	ordersMap := make(map[string]*orderTypes.Order, len(orders))
	for i := 0; i < len(orders); i++ {
		ordersMap[orders[i].ID.String()] = &orders[i]
	}

	for clientID, clientState := range tester.Clients {
		tester.t.Logf("Client %d (%s): order results", clientID, clientState.Address)
		for _, orderState := range clientState.Orders {
			marketExt, err := tester.app.marketKeeper.GetExtended(ctx, orderState.Input.MarketID)
			require.NoError(tester.t, err, "getting marketExt: %s", orderState.Input.MarketID)

			tester.t.Logf("  Order %s (%s):", orderState.ID, orderState.Input.Direction)
			tester.t.Logf("    quantity: %s (%s)", orderState.Input.Quantity, marketExt.BaseCurrency.UintToDec(orderState.Input.Quantity))

			outputOrder, outputExists := ordersMap[orderState.ID.String()]
			if !outputExists {
				tester.t.Logf("    -> fully filled")
			} else {
				tester.t.Logf("    -> partially filled: %s (%s)", outputOrder.Quantity, marketExt.BaseCurrency.UintToDec(outputOrder.Quantity))
			}

			if !orderState.Output.Check {
				continue
			}

			if orderState.Output.FullyFilled {
				require.False(tester.t, outputExists, "order %s: not fully filled", orderState.ID)
			} else {
				require.True(tester.t, outputExists, "order %s: not exists", orderState.ID)
				require.True(tester.t, orderState.Output.PartialQuantity.Equal(outputOrder.Quantity), "order %s: not partially filled", orderState.ID)
				//require.True(tester.t, outputOrder.UpdatedAt.After(orderState.Input.CreatedAt))
			}
		}
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

func Test_OB_BasicNoDecimalAssets(t *testing.T) {
	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(0), uint8(0)

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
		tester.AddClient(client0Addr, sdk.NewInt(1000))
		tester.AddClient(client1Addr, sdk.NewInt(1000))

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

	tester.CheckClientsOutput()
	tester.CheckOrdersOutput()
}

func Test_OB_BasicDiffDecimalAssets(t *testing.T) {
	baseDenom, quoteDenom := "base", "quote"
	baseDecimals, quoteDecimals := uint8(1), uint8(3)

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
		tester.AddClient(client0Addr, sdk.NewInt(1000))
		tester.AddClient(client1Addr, sdk.NewInt(1000))

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

	tester.CheckClientsOutput()
	tester.CheckOrdersOutput()
}
