package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core/msmodule"
)

// Client message to submit a new call.
type MsgSubmitCall struct {
	// Call multi signature message
	Msg msmodule.MsMsg `json:"msg" yaml:"msg"`
	// Call unique ID
	UniqueID string `json:"unique_id" yaml:"unique_id"`
	// Call creator address
	Creator sdk.AccAddress `json:"creator" yaml:"creator"`
}

// Implements sdk.Msg interface.
func (msg MsgSubmitCall) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgSubmitCall) Type() string {
	return "submit_call"
}

// Implements sdk.Msg interface.
func (msg MsgSubmitCall) ValidateBasic() error {
	if msg.UniqueID == "" {
		return sdkErrors.Wrap(ErrWrongCallUniqueId, "empty")
	}

	if msg.Creator.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "creator: empty")
	}

	if err := msg.Msg.ValidateBasic(); err != nil {
		return sdkErrors.Wrap(ErrWrongMsg, err.Error())
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgSubmitCall) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

// Implements sdk.Msg interface.
func (msg MsgSubmitCall) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}

// NewMsgSubmitCall creates a new MsgSubmitCall message.
func NewMsgSubmitCall(msg msmodule.MsMsg, uniqueID string, creator sdk.AccAddress) MsgSubmitCall {
	return MsgSubmitCall{
		UniqueID: uniqueID,
		Msg:      msg,
		Creator:  creator,
	}
}
