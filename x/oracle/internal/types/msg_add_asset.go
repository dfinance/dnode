package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgPostPrice struct representing a adding asset message.
type MsgAddAsset struct {
	// Nominee address
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	// Asset object
	Asset Asset `json:"asset" yaml:"asset"`
}

// Route Implements Msg.
func (msg MsgAddAsset) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddAsset) Type() string { return "add_asset" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgAddAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgAddAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// NewMsgAddAsset creates a new AddAsset message.
func NewMsgAddAsset(
	nominee sdk.AccAddress,
	asset Asset,
) MsgAddAsset {
	return MsgAddAsset{
		Asset:   asset,
		Nominee: nominee,
	}
}
