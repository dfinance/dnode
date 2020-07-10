package types

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CurrentPrice struct that contains the metadata of a current price for a particular asset in the oracle module.
type CurrentPrice struct {
	AssetCode  string    `json:"asset_code" yaml:"asset_code" example:"dfi"` // Denom
	Price      sdk.Int   `json:"price" yaml:"price" swaggertype:"string" example:"1000"`
	ReceivedAt time.Time `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"` // Timestamp Price createdAt
}

// String implement fmt.Stringer for the CurrentPrice.
func (cp CurrentPrice) String() string {
	return strings.TrimSpace(fmt.Sprintf("AssetCode: %s\nPrice: %s\nReceivedAt: %s", cp.AssetCode, cp.Price, cp.ReceivedAt))
}

// PostedPrice struct represented a price for an asset posted by a specific oracle.
type PostedPrice struct {
	AssetCode     string         `json:"asset_code" yaml:"asset_code" example:"dfi"`                                                                        // Denom
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address" swaggertype:"string" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"` // Price source
	Price         sdk.Int        `json:"price" yaml:"price" swaggertype:"string" example:"1000"`
	ReceivedAt    time.Time      `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"` // Timestamp Price createdAt
}

// String implement fmt.Stringer for the PostedPrice.
func (pp PostedPrice) String() string {
	return strings.TrimSpace(
		fmt.Sprintf(
			"AssetCode: %s\nOracleAddress: %s\nPrice: %s\nReceivedAt: %s",
			pp.AssetCode,
			pp.OracleAddress,
			pp.Price,
			pp.ReceivedAt,
		),
	)
}
