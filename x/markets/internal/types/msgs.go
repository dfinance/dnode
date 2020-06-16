package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = MsgCreateMarket{}
)

// Client message to create a market object.
type MsgCreateMarket struct {
	From            sdk.AccAddress `json:"from" yaml:"from"`
	BaseAssetDenom  string         `json:"base_asset_denom" yaml:"base_asset_denom"`
	QuoteAssetDenom string         `json:"quote_asset_denom" yaml:"quote_asset_denom"`
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) Route() string {
	return ModuleName
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) Type() string {
	return "createMarket"
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) ValidateBasic() error {
	if msg.From.Empty() {
		return ErrWrongFrom
	}
	if msg.BaseAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "BaseAsset is empty")
	}
	if msg.QuoteAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "QuoteAsset is empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// NewMsgCreateMarket creates MsgCreateMarket message object.
func NewMsgCreateMarket(fromAddress sdk.AccAddress, baseAsset string, quoteAsset string) MsgCreateMarket {
	return MsgCreateMarket{
		From:            fromAddress,
		BaseAssetDenom:  baseAsset,
		QuoteAssetDenom: quoteAsset,
	}
}
