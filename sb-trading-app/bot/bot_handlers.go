package bot

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func findEventAttrOrderID(prefix string, event coreTypes.ResultEvent) ([]dnTypes.ID, bool) {
	ids := make([]dnTypes.ID, 0)

	orderIdType := prefix + "order_id"
	for eventType, eventValues := range event.Events {
		if eventType == orderIdType {
			if len(eventValues) == 0 {
				return []dnTypes.ID{}, false
			}

			for _, idStr := range eventValues {
				id, err := dnTypes.NewIDFromString(idStr)
				if err != nil {
					return []dnTypes.ID{}, false
				}

				ids = append(ids, id)
			}

			return ids, true
		}
	}

	return []dnTypes.ID{}, false
}

func findEventAttrClearancePrice(event coreTypes.ResultEvent) (sdk.Uint, bool) {
	for eventType, eventValues := range event.Events {
		if eventType == "orderbook.clearance.price" {
			if len(eventValues) != 1 {
				return sdk.Uint{}, false
			}

			return sdk.NewUintFromString(eventValues[0]), true
		}
	}

	return sdk.Uint{}, false
}

func (b *Bot) handleOrderPost(event coreTypes.ResultEvent) {
	orderIDs, ok := findEventAttrOrderID("orders.post.", event)
 	require.True(b.cfg.T, ok, "order_id not found: %v", event)

	for _, orderID := range orderIDs {
		q, order := b.cfg.Tester.QueryOrdersOrder(orderID)
		err := executeQuery(q)
		if err != nil {
			if strings.Contains(err.Error(), "wrong orderID") {
				continue
			}

			require.NoError(b.cfg.T, executeQuery(q), "OrdersOrder on handleOrderPost")
		}

		b.Lock()
		b.orders[order.ID.String()] = *order
		b.logger.Debug(fmt.Sprintf("event: %q order (%s): posted: %s -> %s", order.ID, order.Direction, order.Price, order.Quantity))
		b.Unlock()
	}
}

func (b *Bot) handleOrderCancel(event coreTypes.ResultEvent) {
	orderIDs, ok := findEventAttrOrderID("orders.cancel.", event)
	require.True(b.cfg.T, ok, "order_id not found: %v", event)

	for _, orderID := range orderIDs {
		b.Lock()
		order := b.orders[orderID.String()]
		b.logger.Debug(fmt.Sprintf("event: %q order (%s): canceled", order.ID, order.Direction))
		delete(b.orders, orderID.String())
		b.Unlock()
	}

	b.onOrderCloseMarketMakeMaking("order(s) cancel event")
}

func (b *Bot) handleOrderFullFill(event coreTypes.ResultEvent) {
	orderIDs, ok := findEventAttrOrderID("orders.full_fill.", event)
	require.True(b.cfg.T, ok, "order_id not found: %v", event)

	for _, orderID := range orderIDs {
		b.Lock()
		order := b.orders[orderID.String()]
		b.logger.Debug(fmt.Sprintf("event: %q order (%s): fully filled", order.ID, order.Direction))
		delete(b.orders, orderID.String())
		b.Unlock()
	}

	b.onOrderCloseMarketMakeMaking("order(s) full fill event")
}

func (b *Bot) handleOrderPartialFill(event coreTypes.ResultEvent) {
	orderIDs, ok := findEventAttrOrderID("orders.partial_fill.", event)
	require.True(b.cfg.T, ok, "order_id not found: %v", event)

	for _, orderID := range orderIDs {
		q, order := b.cfg.Tester.QueryOrdersOrder(orderID)
		err := executeQuery(q)
		if err != nil {
			if strings.Contains(err.Error(), "wrong orderID") {
				continue
			}

			require.NoError(b.cfg.T, executeQuery(q), "OrdersOrder on handleOrderPartialFill")
		}

		b.Lock()
		prevOrder := b.orders[order.ID.String()]
		b.orders[order.ID.String()] = *order
		b.logger.Debug(fmt.Sprintf("event: %q order (%s): partially filled (%s / %s)", order.ID, order.Direction, order.Quantity, prevOrder.Quantity))
		b.Unlock()
	}

	b.onOrderCloseMarketMakeMaking("order(s) partially fill event")
}

func (b *Bot) handleOrderBookClearance(event coreTypes.ResultEvent) {
	price, ok := findEventAttrClearancePrice(event)
	require.True(b.cfg.T, ok, "price not found: %v", event)

	b.Lock()
	b.marketPrice = price
	b.logger.Debug(fmt.Sprintf("event: marketPrice updated: %s", price))
	b.Unlock()

	b.onMarketPriceChangeMarketMaking()
}
