package rest

import (
	"github.com/dfinance/dnode/x/oracle"
)

//nolint:deadcode,unused
type (
	OracleRespGetRawPrices struct {
		Height int64                `json:"height"`
		Result []oracle.PostedPrice `json:"result"`
	}

	OracleRespGetPrice struct {
		Height int64               `json:"height"`
		Result oracle.CurrentPrice `json:"result"`
	}

	OracleRespGetAssets struct {
		Height int64         `json:"height"`
		Result oracle.Assets `json:"result"`
	}
)
