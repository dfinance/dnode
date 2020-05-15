package rest

import (
	"github.com/dfinance/dnode/x/markets/internal/types"
)

//nolint:deadcode,unused
type (
	MarketsRespGetMarkets struct {
		Height int64         `json:"height"`
		Result types.Markets `json:"result"`
	}

	MarketsRespGetMarket struct {
		Height int64        `json:"height"`
		Result types.Market `json:"result"`
	}
)
