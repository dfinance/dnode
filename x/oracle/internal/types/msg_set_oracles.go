package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MsgSetOracle struct representing a new nominee based oracle.
type MsgSetOracles struct {
	Oracles Oracles        `json:"oracles" yaml:"oracles"`
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetOracles) ValidateBasic() error {
	if len(msg.Oracles) == 0 {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty oracle addresses array")
	}

	if msg.Denom == "" {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidCoins, "empty denom")
	}

	if msg.Nominee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty nominee")
	}
	return nil
}

// Route Implements Msg.
func (msg MsgSetOracles) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSetOracles) Type() string { return "set_oracles" }

// GetSignBytes Implements Msg.
func (msg MsgSetOracles) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgSetOracles) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// MsgAddOracle creates a new SetOracle message.
func NewMsgSetOracles(
	nominee sdk.AccAddress,
	denom string,
	oracles Oracles,
) MsgSetOracles {
	return MsgSetOracles{
		Oracles: oracles,
		Denom:   denom,
		Nominee: nominee,
	}
}
