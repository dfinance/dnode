package rest

import (
	"github.com/dfinance/dnode/x/pricefeed/internal/types"
)

//nolint:deadcode,unused
type (
	OracleRespGetRawPrices struct {
		Height int64               `json:"height"`
		Result []types.PostedPrice `json:"result"`
	}

	OracleRespGetPrice struct {
		Height int64              `json:"height"`
		Result types.CurrentPrice `json:"result"`
	}

	OracleRespGetAssets struct {
		Height int64        `json:"height"`
		Result types.Assets `json:"result"`
	}
)
