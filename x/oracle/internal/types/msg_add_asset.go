package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Client message to add a new asset.
type MsgAddAsset struct {
	// Nominee address
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	// Asset object
	Asset Asset `json:"asset" yaml:"asset"`
}

// Implements sdk.Msg interface.
func (msg MsgAddAsset) Route() string { return RouterKey }

// Implements sdk.Msg interface.
func (msg MsgAddAsset) Type() string { return "add_asset" }

// Implements sdk.Msg interface.
func (msg MsgAddAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// Implements sdk.Msg interface.
func (msg MsgAddAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg interface.
func (msg MsgAddAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// NewMsgAddAsset creates a new AddAsset message.
func NewMsgAddAsset(nominee sdk.AccAddress, asset Asset, ) MsgAddAsset {
	return MsgAddAsset{
		Asset:   asset,
		Nominee: nominee,
	}
}
