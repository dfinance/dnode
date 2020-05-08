package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"

	orderTypes "github.com/dfinance/dnode/x/order"
)

type OrderAggregate struct {
	Price    sdk.Uint
	Quantity sdk.Uint
}

type OrderAggregates []OrderAggregate

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

func NewBidOrderAggregates(orders orderTypes.Orders) OrderAggregates {
	aggs := make(OrderAggregates, 0, len(orders))
	lastIdx := len(orders) - 1
	if lastIdx <= 0 {
		return aggs
	}

	aggs = append(aggs, OrderAggregate{
		Price:    orders[lastIdx].Price,
		Quantity: orders[lastIdx].Quantity},
	)
	for i := len(orders) - 2; i >= 0; i-- {
		order := &orders[i]

		if aggs[0].Price.Equal(order.Price) {
			aggs[0].Quantity = aggs[0].Quantity.Add(order.Quantity)
			continue
		}

		aggs = append(OrderAggregates{{
			Price:    order.Price,
			Quantity: aggs[0].Quantity.Add(order.Quantity),
		}}, aggs...)
	}

	return aggs
}

func NewAskOrderAggregates(orders orderTypes.Orders) OrderAggregates {
	aggs := make(OrderAggregates, 0, len(orders))
	if len(orders) < 1 {
		return aggs
	}

	aggs = append(aggs, OrderAggregate{
		Price:    orders[0].Price,
		Quantity: orders[0].Quantity},
	)
	for i := 1; i < len(orders); i++ {
		order := &orders[i]
		lastIdx := len(aggs) - 1

		if aggs[lastIdx].Price.Equal(order.Price) {
			aggs[lastIdx].Quantity = aggs[lastIdx].Quantity.Add(order.Quantity)
			continue
		}

		aggs = append(aggs, OrderAggregate{
			Price:    order.Price,
			Quantity: aggs[lastIdx].Quantity.Add(order.Quantity),
		})
	}

	return aggs
}
