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
	BaseAssetDenom  string `json:"base_asset_denom" yaml:"base_asset_denom"`
	QuoteAssetDenom string `json:"quote_asset_denom" yaml:"quote_asset_denom"`
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
	if msg.BaseAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "BaseAsset")
	}
	if msg.QuoteAssetDenom == "" {
		return sdkErrors.Wrap(ErrWrongAssetDenom, "QuoteAsset")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgCreateMarket) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

// NewMsgCreateMarket creates MsgCreateMarket message object.
func NewMsgCreateMarket(baseAsset string, quoteAsset string) MsgCreateMarket {
	return MsgCreateMarket{
		BaseAssetDenom:  baseAsset,
		QuoteAssetDenom: quoteAsset,
	}
}
