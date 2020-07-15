package types

import (
	"time"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Client message to post rawPrice from oracle.
type MsgPostPrice struct {
	// Oracle address
	From sdk.AccAddress `json:"from" yaml:"from"`
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code"`
	// RawPrice
	Price sdk.Int `json:"price" yaml:"price"`
	// ReceivedAt time in UNIX timestamp format [seconds]
	ReceivedAt time.Time `json:"received_at" yaml:"received_at"`
}

// Implements sdk.Msg interface.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Implements sdk.Msg interface.
func (msg MsgPostPrice) Type() string { return "post_price" }

// Implements sdk.Msg interface.
func (msg MsgPostPrice) ValidateBasic() error {
	if msg.From.Empty() {
		return sdkErrors.Wrap(ErrInternal, "invalid (empty) oracle address")
	}
	if err := msg.AssetCode.Validate(); err != nil {
		return sdkErrors.Wrapf(ErrInternal, "invalid assetCode: value (%s), error (%v)", msg.AssetCode, err)
	}
	if msg.Price.IsNegative() {
		return sdkErrors.Wrap(ErrInternal, "invalid (negative) price")
	}
	if msg.Price.BigInt().BitLen() > PriceBytesLimit*8 {
		return sdkErrors.Wrapf(ErrInternal, "out of %d bytes limit for price", PriceBytesLimit)
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg interface.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// NewMsgPostPrice creates a new PostPrice message.
func NewMsgPostPrice(from sdk.AccAddress, assetCode dnTypes.AssetCode, price sdk.Int, receivedAt time.Time) MsgPostPrice {
	return MsgPostPrice{
		From:       from,
		AssetCode:  assetCode,
		Price:      price,
		ReceivedAt: receivedAt,
	}
}
