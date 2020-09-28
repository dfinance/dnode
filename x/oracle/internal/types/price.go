package types

import (
	"fmt"
	"github.com/shopspring/decimal"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	PriceBytesLimit = 8
	PricePrecision  = 8
)

// CurrentPrice contains meta of the current price for the particular asset with ask and bid prices.
type CurrentPrice struct {
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code" example:"btc_xfi"`
	// AskPrice
	AskPrice sdk.Int `json:"ask_price" yaml:"ask_price" swaggertype:"string" example:"1000"`
	// BidPrice
	BidPrice sdk.Int `json:"bid_price" yaml:"bid_price" swaggertype:"string" example:"1000"`
	// UNIX Timestamp price createdAt [sec]
	ReceivedAt time.Time `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
}

// GetReversedAssetCurrentPrice returns CurrentPrice for reverted
func (cp CurrentPrice) GetReversedAssetCurrentPrice() CurrentPrice {
	revertInt := func(p sdk.Int) sdk.Int {
		decP := decimal.NewFromBigInt(p.BigInt(), -PricePrecision)
		decP = decimal.NewFromInt(1).Div(decP)
		decP = decP.Mul(decimal.NewFromInt(10).Pow(decimal.NewFromInt(PricePrecision)))
		return sdk.NewIntFromBigInt(decP.BigInt())
	}

	return CurrentPrice{
		AssetCode:  cp.AssetCode.ReverseCode(),
		AskPrice:   revertInt(cp.BidPrice),
		BidPrice:   revertInt(cp.AskPrice),
		ReceivedAt: cp.ReceivedAt,
	}
}

// CurrentAssetPrice contains meta of the current price for the particular asset.
type CurrentAssetPrice struct {
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code" example:"btc_xfi"`
	// Price
	Price sdk.Int `json:"price" yaml:"price" swaggertype:"string" example:"1000"`
	// UNIX Timestamp price createdAt [sec]
	ReceivedAt time.Time `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
}

// Valid checks that CurrentPrice is valid (used for genesis ops).
func (cp CurrentPrice) Valid() error {
	if err := cp.AssetCode.Validate(); err != nil {
		return fmt.Errorf("asset_code: %w", err)
	}
	if cp.AskPrice.IsZero() {
		return fmt.Errorf("askPrice: is zero")
	}
	if cp.AskPrice.IsNegative() {
		return fmt.Errorf("askPrice: is negative")
	}
	if cp.BidPrice.IsZero() {
		return fmt.Errorf("bidPrice: is zero")
	}
	if cp.BidPrice.IsNegative() {
		return fmt.Errorf("bidPrice: is negative")
	}
	if cp.ReceivedAt.IsZero() {
		return fmt.Errorf("received_at: is zero")
	}
	return nil
}

func (cp CurrentPrice) String() string {
	return fmt.Sprintf("CurrentPrice:\n"+
		"AssetCode: %s\n"+
		"AskPrice: %s\n"+
		"BidPrice: %s\n"+
		"ReceivedAt: %s",
		cp.AssetCode, cp.AskPrice, cp.BidPrice, cp.ReceivedAt,
	)
}

type CurrentPrices []CurrentPrice

// PostedPrice contains price for an asset posted by a specific oracle.
type PostedPrice struct {
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code" example:"btc_xfi"`
	// Source oracle address
	OracleAddress sdk.AccAddress `json:"oracle_address" yaml:"oracle_address" swaggertype:"string" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// AskPrice
	AskPrice sdk.Int `json:"ask_price" yaml:"ask_price" swaggertype:"string" example:"1000"`
	// BidPrice
	BidPrice sdk.Int `json:"bid_price" yaml:"bid_price" swaggertype:"string" example:"1000"`
	// UNIX Timestamp price receivedAt [sec]
	ReceivedAt time.Time `json:"received_at" yaml:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
}

// String implement fmt.Stringer for the PostedPrice type.
func (pp PostedPrice) String() string {
	return strings.TrimSpace(
		fmt.Sprintf(
			"AssetCode: %s\nOracleAddress: %s\nAskPrice: %s\nBidPrice: %s\nReceivedAt: %s",
			pp.AssetCode,
			pp.OracleAddress,
			pp.AskPrice,
			pp.BidPrice,
			pp.ReceivedAt,
		),
	)
}

// PendingPriceAsset contains info about the asset which price is still to be determined.
type PendingPriceAsset struct {
	AssetCode string `json:"asset_code"`
}

func (a PendingPriceAsset) String() string {
	return fmt.Sprintf(`AssetCode: %s`, a.AssetCode)
}
