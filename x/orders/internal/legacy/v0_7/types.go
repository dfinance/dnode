package v0_7

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orders/internal/legacy/v0_6"
)

type (
	Order struct {
		ID        dnTypes.ID             `json:"id"`
		Owner     sdk.AccAddress         `json:"owner"`
		Market    markets.MarketExtended `json:"market"`
		Direction v0_6.Direction         `json:"direction"`
		Price     sdk.Uint               `json:"price"`
		Quantity  sdk.Uint               `json:"quantity"`
		Ttl       time.Duration          `json:"ttl_dur"`
		Memo      string                 `json:"memo"`
		CreatedAt time.Time              `json:"created_at"`
		UpdatedAt time.Time              `json:"updated_at"`
	}

	Orders []Order

	GenesisState struct {
		Orders      Orders      `json:"orders" yaml:"orders"`
		LastOrderId *dnTypes.ID `json:"last_order_id" yaml:"last_order_id"`
	}
)
