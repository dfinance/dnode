package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Client message to update an existing asset.
type MsgSetAsset struct {
	// Nominee address
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	// Asset object
	Asset Asset `json:"asset" yaml:"asset"`
}

// Implements sdk.Msg interface.
func (msg MsgSetAsset) Route() string { return RouterKey }

// Implements sdk.Msg interface.
func (msg MsgSetAsset) Type() string { return "set_asset" }

// Implements sdk.Msg interface.
func (msg MsgSetAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// Implements sdk.Msg interface.
func (msg MsgSetAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg interface.
func (msg MsgSetAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// NewMsgSetAsset creates a new SetAsset message.
func NewMsgSetAsset(nominee sdk.AccAddress, asset Asset, ) MsgSetAsset {
	return MsgSetAsset{
		Asset:   asset,
		Nominee: nominee,
	}
}
