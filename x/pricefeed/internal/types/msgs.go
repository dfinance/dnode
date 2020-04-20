package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// TypeMsgPostPrice type of PostPrice msg
	TypeMsgPostPrice = "post_price"
)

// MsgPostPrice struct representing a posted price message.
// Used by price feeds to input prices to the price feed
type MsgPostPrice struct {
	From       sdk.AccAddress `json:"from" yaml:"from"`
	AssetCode  string         `json:"asset_code" yaml:"asset_code"`
	Price      sdk.Int        `json:"price" yaml:"price"`
	ReceivedAt time.Time      `json:"received_at" yaml:"received_at"`
}

// NewMsgPostPrice creates a new post price msg
func NewMsgPostPrice(
	from sdk.AccAddress,
	assetCode string,
	price sdk.Int,
	receivedAt time.Time) MsgPostPrice {
	return MsgPostPrice{
		From:       from,
		AssetCode:  assetCode,
		Price:      price,
		ReceivedAt: receivedAt,
	}
}

// Route Implements Msg.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgPostPrice) Type() string { return TypeMsgPostPrice }

// GetSignBytes Implements Msg.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPostPrice) ValidateBasic() error {
	if msg.From.Empty() {
		return sdkErrors.Wrap(ErrInternal, "invalid (empty) price feed address")
	}
	if len(msg.AssetCode) == 0 {
		return sdkErrors.Wrap(ErrInternal, "invalid (empty) asset code")
	}
	if msg.Price.IsNegative() {
		return sdkErrors.Wrap(ErrInternal, "invalid (negative) price")
	}
	if msg.Price.BigInt().BitLen() > PriceBytesLimit*8 {
		return sdkErrors.Wrapf(ErrInternal, "out of %d bytes limit for price", PriceBytesLimit)
	}
	// TODO check coin denoms

	return nil
}

// MsgAddPriceFeed struct representing a new nominee based price feed
type MsgAddPriceFeed struct {
	PriceFeed sdk.AccAddress `json:"price_feed" yaml:"price_feed"`
	Nominee   sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom     string         `json:"denom" yaml:"denom"`
}

// MsgAddPriceFeed creates a new add price feed message
func NewMsgAddPriceFeed(
	nominee sdk.AccAddress,
	denom string,
	pricefeed sdk.AccAddress,
) MsgAddPriceFeed {
	return MsgAddPriceFeed{
		PriceFeed: pricefeed,
		Denom:     denom,
		Nominee:   nominee,
	}
}

// Route Implements Msg.
func (msg MsgAddPriceFeed) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddPriceFeed) Type() string { return "add_pricefeed" }

// GetSignBytes Implements Msg.
func (msg MsgAddPriceFeed) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddPriceFeed) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgAddPriceFeed) ValidateBasic() error {
	if msg.PriceFeed.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty price feed address")
	}

	if msg.Denom == "" {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidCoins, "empty denom")
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}

	return nil
}

// MsgSetOracle struct representing a new nominee based price feed
type MsgSetPriceFeeds struct {
	PriceFeeds PriceFeeds     `json:"price_feeds" yaml:"price_feeds"`
	Nominee    sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom      string         `json:"denom" yaml:"denom"`
}

// MsgAddPriceFeed creates a new add price feed message
func NewMsgSetOracles(
	nominee sdk.AccAddress,
	denom string,
	pricefeeds PriceFeeds,
) MsgSetPriceFeeds {
	return MsgSetPriceFeeds{
		PriceFeeds: pricefeeds,
		Denom:      denom,
		Nominee:    nominee,
	}
}

// Route Implements Msg.
func (msg MsgSetPriceFeeds) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgSetPriceFeeds) Type() string { return "set_pricefeeds" }

// GetSignBytes Implements Msg.
func (msg MsgSetPriceFeeds) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgSetPriceFeeds) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetPriceFeeds) ValidateBasic() error {
	if len(msg.PriceFeeds) == 0 {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty price feed addresses array")
	}

	if msg.Denom == "" {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidCoins, "empty denom")
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// MsgSetAsset struct representing a new nominee based price feed
type MsgSetAsset struct {
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
	Asset   Asset          `json:"asset" yaml:"asset"`
}

// NewMsgSetAsset creat  es a new add price feed message
func NewMsgSetAsset(
	nominee sdk.AccAddress,
	denom string,
	asset Asset,
) MsgSetAsset {
	return MsgSetAsset{
		Asset:   asset,
		Denom:   denom,
		Nominee: nominee,
	}
}

// Route Implements Msg.
func (msg MsgSetAsset) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgSetAsset) Type() string { return "set_asset" }

// GetSignBytes Implements Msg.
func (msg MsgSetAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgSetAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if len(msg.Denom) == 0 {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidCoins, "missing denom")
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

type MsgAddAsset struct {
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
	Asset   Asset          `json:"asset" yaml:"asset"`
}

func NewMsgAddAsset(
	nominee sdk.AccAddress,
	denom string,
	asset Asset,
) MsgAddAsset {
	return MsgAddAsset{
		Asset:   asset,
		Denom:   denom,
		Nominee: nominee,
	}
}

// Route Implements Msg.
func (msg MsgAddAsset) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddAsset) Type() string { return "add_asset" }

// GetSignBytes Implements Msg.
func (msg MsgAddAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgAddAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if msg.Denom == "" {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidCoins, "empty denom")
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}
