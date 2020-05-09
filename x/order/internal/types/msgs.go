package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

var (
	_ sdk.Msg = MsgPostOrder{}
	_ sdk.Msg = MsgRevokeOrder{}
)

// Client message to post an order object.
type MsgPostOrder struct {
	Owner     sdk.AccAddress `json:"owner" yaml:"owner"`
	MarketID  dnTypes.ID     `json:"market_id" yaml:"market_id"`
	Direction Direction      `json:"direction" yaml:"direction"`
	Price     sdk.Uint       `json:"price" yaml:"price"`
	Quantity  sdk.Uint       `json:"quantity" yaml:"quantity"`
	TtlInSec  uint64         `json:"ttl_in_sec" yaml:"ttl_in_sec"`
}

// Implements sdk.Msg interface.
func (msg MsgPostOrder) Route() string {
	return "order"
}

// Implements sdk.Msg interface.
func (msg MsgPostOrder) Type() string {
	return "post"
}

// Implements sdk.Msg interface.
func (msg MsgPostOrder) ValidateBasic() error {
	if err := msg.MarketID.Valid(); err != nil {
		return sdkErrors.Wrap(ErrWrongMarketID, err.Error())
	}
	if msg.Owner.Empty() {
		return ErrWrongOwner
	}
	if !msg.Direction.IsValid() {
		return ErrWrongDirection
	}
	if msg.Price.IsZero() {
		return ErrWrongPrice
	}
	if msg.Quantity.IsZero() {
		return ErrWrongQuantity
	}
	if msg.TtlInSec == 0 {
		return ErrWrongTtl
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgPostOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgPostOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// NewMsgPost creates MsgPostOrder message object.
func NewMsgPost(owner sdk.AccAddress, marketID dnTypes.ID, direction Direction, price sdk.Uint, quantity sdk.Uint, ttlInSec uint64) MsgPostOrder {
	return MsgPostOrder{
		Owner:     owner,
		MarketID:  marketID,
		Direction: direction,
		Price:     price,
		Quantity:  quantity,
		TtlInSec:  ttlInSec,
	}
}

// Client message to revoke an order.
type MsgRevokeOrder struct {
	Owner   sdk.AccAddress `json:"owner" yaml:"owner"`
	OrderID dnTypes.ID     `json:"order_id" yaml:"order_id"`
}

// Implements sdk.Msg interface.
func (msg MsgRevokeOrder) Route() string {
	return "order"
}

// Implements sdk.Msg interface.
func (msg MsgRevokeOrder) Type() string {
	return "cancel"
}

// Implements sdk.Msg interface.
func (msg MsgRevokeOrder) ValidateBasic() error {
	if msg.Owner.Empty() {
		return ErrWrongOwner
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgRevokeOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgRevokeOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

// NewMsgRevokeOrder creates MsgRevokeOrder message object.
func NewMsgRevokeOrder(owner sdk.AccAddress, id dnTypes.ID) MsgRevokeOrder {
	return MsgRevokeOrder{
		Owner:   owner,
		OrderID: id,
	}
}
