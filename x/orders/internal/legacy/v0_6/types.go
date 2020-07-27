package v0_6

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets"
)

type (
	Direction string

	Order struct {
		ID        dnTypes.ID             `json:"id"`
		Owner     sdk.AccAddress         `json:"owner"`
		Market    markets.MarketExtended `json:"market"`
		Direction Direction              `json:"direction"`
		Price     sdk.Uint               `json:"price"`
		Quantity  sdk.Uint               `json:"quantity"`
		Ttl       time.Duration          `json:"ttl_dur"`
		CreatedAt time.Time              `json:"created_at"`
		UpdatedAt time.Time              `json:"updated_at"`
	}

	Orders []Order

	GenesisState struct {
		Orders      Orders      `json:"orders"`
		LastOrderId *dnTypes.ID `json:"last_order_id"`
	}
)
