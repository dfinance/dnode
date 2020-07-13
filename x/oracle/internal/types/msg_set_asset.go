package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSetAsset struct representing a new nominee based oracle.
type MsgSetAsset struct {
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Asset   Asset          `json:"asset" yaml:"asset"`
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetAsset) ValidateBasic() error {
	if err := msg.Asset.ValidateBasic(); err != nil {
		return err
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// Route Implements Msg.
func (msg MsgSetAsset) Route() string { return RouterKey }

// Type Implements Msg.
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

// NewMsgSetAsset creates a new SetAsset message.
func NewMsgSetAsset(
	nominee sdk.AccAddress,
	asset Asset,
) MsgSetAsset {
	return MsgSetAsset{
		Asset:   asset,
		Nominee: nominee,
	}
}
