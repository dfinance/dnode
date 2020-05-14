package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"

	orderTypes "github.com/dfinance/dnode/x/orders"
)

// OrderAggregate type stores aggregated quantity (relative to price) for bid/ask orders.
// Bid/ask aggregates are combined to build PQCurve.
type OrderAggregate struct {
	Price    sdk.Uint
	Quantity sdk.Uint
}

// OrderAggregate sort.Interface.
type OrderAggregates []OrderAggregate

// Strings returns multi-line text object representation.
func (a *OrderAggregates) String() string {
	var buf bytes.Buffer

	t := tablewriter.NewWriter(&buf)
	t.SetHeader([]string{
		"OA.Price",
		"OA.Price",
	})

	for _, o := range *a {
		t.Append([]string{
			o.Price.String(),
			o.Quantity.String(),
		})
	}
	t.Render()

	return string(buf.Bytes())
}

// NewBidOrderAggregates groups bid orders by price summing quantities.
// Contract: orders must be price sorted (ASC).
// Result is price sorted (ASC).
func NewBidOrderAggregates(orders orderTypes.Orders) OrderAggregates {
	aggs := make(OrderAggregates, 0, len(orders))
	lastIdx := len(orders) - 1
	if lastIdx < 0 {
		return aggs
	}

	// add the first element with the highest price
	aggs = append(aggs, OrderAggregate{
		Price:    orders[lastIdx].Price,
		Quantity: orders[lastIdx].Quantity},
	)
	for i := len(orders) - 2; i >= 0; i-- {
		order := &orders[i]

		// increase the aggregate quantity is price already exists
		if aggs[0].Price.Equal(order.Price) {
			aggs[0].Quantity = aggs[0].Quantity.Add(order.Quantity)
			continue
		}

		// prepend the aggregate if price wasn't not found
		aggs = append(OrderAggregates{{
			Price:    order.Price,
			Quantity: aggs[0].Quantity.Add(order.Quantity),
		}}, aggs...)
	}

	return aggs
}

// NewAskOrderAggregates groups ask orders by price summing quantities.
// Contract: orders must be price sorted (ASC).
// Result is price sorted (ASC).
func NewAskOrderAggregates(orders orderTypes.Orders) OrderAggregates {
	aggs := make(OrderAggregates, 0, len(orders))
	if len(orders) < 1 {
		return aggs
	}

	// add the first element with the lowest price
	aggs = append(aggs, OrderAggregate{
		Price:    orders[0].Price,
		Quantity: orders[0].Quantity},
	)
	for i := 1; i < len(orders); i++ {
		order := &orders[i]
		lastIdx := len(aggs) - 1

		// increase the aggregate quantity is price already exists
		if aggs[lastIdx].Price.Equal(order.Price) {
			aggs[lastIdx].Quantity = aggs[lastIdx].Quantity.Add(order.Quantity)
			continue
		}

		// append the aggregate if price wasn't not found
		aggs = append(aggs, OrderAggregate{
			Price:    order.Price,
			Quantity: aggs[lastIdx].Quantity.Add(order.Quantity),
		})
	}

	return aggs
}
