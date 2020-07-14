package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// MsgAddOracle struct representing a new nominee based oracle.
type MsgAddOracle struct {
	Oracle    sdk.AccAddress    `json:"oracle" yaml:"oracle"`
	Nominee   sdk.AccAddress    `json:"nominee" yaml:"nominee"`
	AssetCode dnTypes.AssetCode `json:"asset_code" yaml:"asset_code"`
}

// Route Implements Msg.
func (msg MsgAddOracle) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddOracle) Type() string { return "add_oracle" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
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

// GetSignBytes Implements Msg.
func (msg MsgAddOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)

	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddOracle) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// MsgAddOracle creates a new AddOracle message.
func NewMsgAddOracle(
	nominee sdk.AccAddress,
	assetCode dnTypes.AssetCode,
	oracle sdk.AccAddress,
) MsgAddOracle {
	return MsgAddOracle{
		Oracle:    oracle,
		AssetCode: assetCode,
		Nominee:   nominee,
	}
}
