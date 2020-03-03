package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// TypeMsgPostPrice type of PostPrice msg
	TypeMsgPostPrice = "post_price"
)

// MsgPostPrice struct representing a posted price message.
// Used by oracles to input prices to the oracle
type MsgPostPrice struct {
	From       sdk.AccAddress `json:"from" yaml:"from"`
	AssetCode  string         `json:"asset_code" yaml:"asset_code"`
	Price      sdk.Int        `json:"price" yaml:"price"`
	ReceivedAt time.Time      `json:"received_at" yaml:"received_at"`
}

// NewMsgPostPrice creates a new post price msg
func NewMsgPostPrice(
	from sdk.AccAddress,
	assetCode string,
	price sdk.Int,
	receivedAt time.Time) MsgPostPrice {
	return MsgPostPrice{
		From:       from,
		AssetCode:  assetCode,
		Price:      price,
		ReceivedAt: receivedAt,
	}
}

// Route Implements Msg.
func (msg MsgPostPrice) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgPostPrice) Type() string { return TypeMsgPostPrice }

// GetSignBytes Implements Msg.
func (msg MsgPostPrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgPostPrice) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPostPrice) ValidateBasic() sdk.Error {
	if msg.From.Empty() {
		return sdk.ErrInternal("invalid (empty) oracle address")
	}
	if len(msg.AssetCode) == 0 {
		return sdk.ErrInternal("invalid (empty) asset code")
	}
	if msg.Price.IsNegative() {
		return sdk.ErrInternal("invalid (negative) price")
	}
	if msg.Price.BigInt().BitLen() > PriceBytesLimit*8 {
		return sdk.ErrInternal(fmt.Sprintf("out of %d bytes limit for price", PriceBytesLimit))
	}
	// TODO check coin denoms
	return nil
}

// MsgAddOracle struct representing a new nominee based oracle
type MsgAddOracle struct {
	Oracle  sdk.AccAddress `json:"oracle" yaml:"oracle"`
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
}

// MsgAddOracle creates a new add oracle message
func NewMsgAddOracle(
	nominee sdk.AccAddress,
	denom string,
	oracle sdk.AccAddress,
) MsgAddOracle {
	return MsgAddOracle{
		Oracle:  oracle,
		Denom:   denom,
		Nominee: nominee,
	}
}

// Route Implements Msg.
func (msg MsgAddOracle) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddOracle) Type() string { return "add_oracle" }

// GetSignBytes Implements Msg.
func (msg MsgAddOracle) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddOracle) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgAddOracle) ValidateBasic() sdk.Error {
	if msg.Oracle.Empty() {
		return sdk.ErrInvalidAddress("missing oracle address")
	}

	if msg.Denom == "" {
		return sdk.ErrInvalidCoins("missing denom")
	}

	if msg.Nominee.Empty() {
		return sdk.ErrInvalidAddress("missing nominee address")
	}
	return nil
}

// MsgSetOracle struct representing a new nominee based oracle
type MsgSetOracles struct {
	Oracles Oracles        `json:"oracles" yaml:"oracles"`
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
}

// MsgAddOracle creates a new add oracle message
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

// Route Implements Msg.
func (msg MsgSetOracles) Route() string { return RouterKey }

// Type Implements Msg
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

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetOracles) ValidateBasic() sdk.Error {
	if len(msg.Oracles) == 0 {
		return sdk.ErrInvalidAddress("missing oracle address")
	}

	if msg.Denom == "" {
		return sdk.ErrInvalidCoins("missing denom")
	}

	if msg.Nominee.Empty() {
		return sdk.ErrInvalidAddress("missing nominee address")
	}
	return nil
}

// MsgSetAsset struct representing a new nominee based oracle
type MsgSetAsset struct {
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
	Asset   Asset          `json:"asset" yaml:"asset"`
}

// NewMsgSetAsset creates a new add oracle message
func NewMsgSetAsset(
	nominee sdk.AccAddress,
	denom string,
	asset Asset,
) MsgSetAsset {
	return MsgSetAsset{
		Asset:   asset,
		Denom:   denom,
		Nominee: nominee,
	}
}

// Route Implements Msg.
func (msg MsgSetAsset) Route() string { return RouterKey }

// Type Implements Msg
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

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSetAsset) ValidateBasic() sdk.Error {
	err := msg.Asset.ValidateBasic()
	if err != nil {
		return err
	}

	if len(msg.Denom) == 0 {
		return sdk.ErrInvalidCoins("missing denom")
	}

	if msg.Nominee.Empty() {
		return sdk.ErrInvalidAddress("missing nominee address")
	}
	return nil
}

type MsgAddAsset struct {
	Nominee sdk.AccAddress `json:"nominee" yaml:"nominee"`
	Denom   string         `json:"denom" yaml:"denom"`
	Asset   Asset          `json:"asset" yaml:"asset"`
}

func NewMsgAddAsset(
	nominee sdk.AccAddress,
	denom string,
	asset Asset,
) MsgAddAsset {
	return MsgAddAsset{
		Asset:   asset,
		Denom:   denom,
		Nominee: nominee,
	}
}

// Route Implements Msg.
func (msg MsgAddAsset) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgAddAsset) Type() string { return "add_asset" }

// GetSignBytes Implements Msg.
func (msg MsgAddAsset) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners Implements Msg.
func (msg MsgAddAsset) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Nominee}
}

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgAddAsset) ValidateBasic() sdk.Error {
	err := msg.Asset.ValidateBasic()
	if err != nil {
		return err
	}

	if msg.Denom == "" {
		return sdk.ErrInvalidCoins("missing denom")
	}

	if msg.Nominee.Empty() {
		return sdk.ErrInvalidAddress("missing nominee address")
	}
	return nil
}
