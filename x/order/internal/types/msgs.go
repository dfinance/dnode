package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

var (
	_ sdk.Msg = MsgPostOrder{}
	_ sdk.Msg = MsgCancelOrder{}
)

type MsgPostOrder struct {
	Owner     sdk.AccAddress `json:"owner" yaml:"owner"`
	MarketID  dnTypes.ID     `json:"market_id" yaml:"market_id"`
	Direction Direction      `json:"direction" yaml:"direction"`
	Price     sdk.Uint       `json:"price" yaml:"price"`
	Quantity  sdk.Uint       `json:"quantity" yaml:"quantity"`
	TtlInSec  uint64         `json:"ttl_in_sec" yaml:"ttl_in_sec"`
}

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

func (msg MsgPostOrder) Route() string {
	return "order"
}

func (msg MsgPostOrder) Type() string {
	return "post"
}

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

func (msg MsgPostOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgPostOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}

type MsgCancelOrder struct {
	Owner   sdk.AccAddress `json:"owner" yaml:"owner"`
	OrderID dnTypes.ID     `json:"order_id" yaml:"order_id"`
}

func NewMsgCancelOrder(owner sdk.AccAddress, id dnTypes.ID) MsgCancelOrder {
	return MsgCancelOrder{
		Owner:   owner,
		OrderID: id,
	}
}

func (msg MsgCancelOrder) Route() string {
	return "order"
}

func (msg MsgCancelOrder) Type() string {
	return "cancel"
}

func (msg MsgCancelOrder) ValidateBasic() error {
	if msg.Owner.Empty() {
		return ErrWrongOwner
	}

	return nil
}

func (msg MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg MsgCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Owner}
}
