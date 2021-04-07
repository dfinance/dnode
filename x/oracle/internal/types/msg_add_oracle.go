package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to add oracle source to an existing asset.
type MsgAddOracle struct {
	// Oracle address
	Oracle sdk.AccAddress `json:"oracle" yaml:"oracle"`
	// Nominee address
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	// Asset code
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code"`
}

// Implements sdk.Msg interface.
func (msg MsgAddOracle) Route() string { return RouterKey }

// Implements sdk.Msg interface.
func (msg MsgAddOracle) Type() string { return "add_oracle" }

// Implements sdk.Msg interface.
func (msg MsgAddOracle) ValidateBasic() error {
	if msg.Oracle.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty oracle address")
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
func (msg MsgAddOracle) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgAddOracle) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// MsgAddOracle creates a new AddOracle message.
func NewMsgAddOracle(nominee sdk.AccAddress, assetCode dnTypes.AssetCode, oracle sdk.AccAddress) MsgAddOracle {
	return MsgAddOracle{
		Oracle:    oracle,
		AssetCode: assetCode,
		Nominee:   nominee,
	}
}
