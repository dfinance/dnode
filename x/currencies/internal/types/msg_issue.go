package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client multisig message to issue currency.
type MsgIssueCurrency struct {
	// Issue unique ID (could be txHash of transaction in another blockchain)
	ID string `json:"id"`
	// Target currency issue coin
	Coin sdk.Coin `json:"coin"`
	// Target account
	Payee sdk.AccAddress `json:"Payee"`
}

// Implements sdk.Msg interface.
func (msg MsgIssueCurrency) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgIssueCurrency) Type() string {
	return "issue_currency"
}

// Implements sdk.Msg interface.
func (msg MsgIssueCurrency) ValidateBasic() error {
	if len(msg.ID) == 0 {
		return sdkErrors.Wrap(ErrWrongIssueID, "empty")
	}

	if err := dnTypes.DenomFilter(msg.Coin.Denom); err != nil {
		return sdkErrors.Wrap(ErrWrongDenom, err.Error())
	}

	if msg.Coin.Amount.LTE(sdk.ZeroInt()) {
		return sdkErrors.Wrap(ErrWrongAmount, "LTE to zero")
	}

	if msg.Payee.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "payee: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgIssueCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
// Msg is a multisig, so there are not signers.
func (msg MsgIssueCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

// NewMsgIssueCurrency creates a new MsgIssueCurrency message.
func NewMsgIssueCurrency(id string, coin sdk.Coin, payee sdk.AccAddress) MsgIssueCurrency {
	return MsgIssueCurrency{
		ID:    id,
		Coin:  coin,
		Payee: payee,
	}
}
