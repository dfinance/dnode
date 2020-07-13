package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgPostPrice struct representing a posted price message.
// Used by oracles to input prices to the oracle
type MsgPostPrice struct {
	From       sdk.AccAddress    `json:"from" yaml:"from"`
	AssetCode  dnTypes.AssetCode `json:"asset_code" yaml:"asset_code"`
	Price      sdk.Int           `json:"price" yaml:"price"`
	ReceivedAt time.Time         `json:"received_at" yaml:"received_at"`
}

// Route Implements Msg.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgPostPrice) Type() string { return "post_price" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPostPrice) ValidateBasic() error {
	if msg.From.Empty() {
		return sdkErrors.Wrap(ErrInternal, "invalid (empty) oracle address")
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

	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// NewMsgPostPrice creates a new PostPrice message.
func NewMsgPostPrice(
	from sdk.AccAddress,
	assetCode dnTypes.AssetCode,
	price sdk.Int,
	receivedAt time.Time) MsgPostPrice {
	return MsgPostPrice{
		From:       from,
		AssetCode:  assetCode,
		Price:      price,
		ReceivedAt: receivedAt,
	}
}
