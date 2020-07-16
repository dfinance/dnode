package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to set oracle source for an existing asset.
type MsgSetOracles struct {
	// Array of oracles addresses
	Oracles Oracles `json:"oracles" yaml:"oracles"`
	// Nominee address
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code"`
}

// Implements sdk.Msg interface.
func (msg MsgSetOracles) Route() string { return RouterKey }

// Implements sdk.Msg interface.
func (msg MsgSetOracles) Type() string { return "set_oracles" }

// Implements sdk.Msg interface.
func (msg MsgSetOracles) ValidateBasic() error {
	if len(msg.Oracles) == 0 {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty oracle addresses array")
	}

	if err := msg.AssetCode.Validate(); err != nil {
		return sdkErrors.Wrapf(ErrInternal, "invalid assetCode: value (%s), error (%v)", msg.AssetCode, err)
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// Implements sdk.Msg interface.
func (msg MsgSetOracles) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// Implements sdk.Msg interface.
func (msg MsgSetOracles) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// MsgAddOracle creates a new SetOracle message.
func NewMsgSetOracles(nominee sdk.AccAddress, assetCode dnTypes.AssetCode, oracles Oracles, ) MsgSetOracles {
	return MsgSetOracles{
		Oracles:   oracles,
		AssetCode: assetCode,
		Nominee:   nominee,
	}
}
