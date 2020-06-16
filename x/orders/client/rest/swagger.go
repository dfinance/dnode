package rest

import (
	"github.com/dfinance/dnode/x/orders/internal/types"
)

//nolint:deadcode,unused
type (
	OrdersRespGetOrders struct {
		Height int64        `json:"height"`
		Result types.Orders `json:"result"`
	}

	OrdersRespGetOrder struct {
		Height int64       `json:"height"`
		Result types.Order `json:"result"`
	}
)
