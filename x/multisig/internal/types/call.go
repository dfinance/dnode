package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
)

// Call contains multi signature message with some meta.
type Call struct {
	// Call ID
	ID dnTypes.ID `json:"id" yaml:"id" example:"0" format:"string representation for big.Uint" swaggertype:"string"`
	// Call unique ID (ID and UniqueID both identifies call)
	UniqueID string `json:"unique_id" yaml:"unique_id" example:"issue1"`
	// Call creator address
	Creator sdk.AccAddress `json:"creator" yaml:"creator" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Call state: approved to execute
	Approved bool `json:"approved" yaml:"approved"`
	// Call state: executed
	Executed bool `json:"executed" yaml:"executed"`
	// Call state: execution failed
	Failed bool `json:"failed" yaml:"failed"`
	// Call state: rejected
	Rejected bool `json:"rejected" yaml:"rejected"`
	// Call fail reason
	Error string `json:"error" yaml:"error"`
	// Message: data
	Msg msmodule.MsMsg `json:"msg_data" yaml:"msg_data"`
	// Message: route
	MsgRoute string `json:"msg_route" yaml:"msg_route" example:"oracle"`
	// Message: type
	MsgType string `json:"msg_type" yaml:"msg_type" example:"add_asset"`
	// BlockHeight when call was submitted
	Height int64 `json:"height" yaml:"height" example:"1"`
}

// CanBeVoted checks if call accepts votes (vote / revoke confirmation).
func (c Call) CanBeVoted() error {
	if c.Approved {
		return sdkErrors.Wrap(ErrVoteAlreadyApproved, c.ID.String())
	}
	if c.Rejected {
		return sdkErrors.Wrap(ErrVoteAlreadyRejected, c.ID.String())
	}

	return nil
}

func (c Call) String() string {
	return fmt.Sprintf("Call:\n"+
		"  ID:       %s\n"+
		"  UniqueID: %s\n"+
		"  Creator:  %s\n"+
		"  Approved: %v\n"+
		"  Executed: %v\n"+
		"  Failed:   %v\n"+
		"  Rejected: %v\n"+
		"  Error:    %s\n"+
		"  MsgRoute: %s\n"+
		"  MsgType:  %s\n"+
		"  Height:   %d",
		c.ID.String(),
		c.UniqueID,
		c.Creator.String(),
		c.Approved,
		c.Executed,
		c.Failed,
		c.Rejected,
		c.Error,
		c.MsgRoute,
		c.MsgType,
		c.Height,
	)
}

// NewCall creates a new Call object.
func NewCall(id dnTypes.ID, uniqueID string, msg msmodule.MsMsg, blockHeight int64, creatorAddr sdk.AccAddress) (Call, error) {
	// check message
	if msg == nil {
		return Call{}, sdkErrors.Wrap(ErrWrongMsg, "nil")
	}

	msgRoute := msg.Route()
	if msgRoute == "" {
		return Call{}, sdkErrors.Wrap(ErrWrongMsgRoute, "empty")
	}

	msgType := msg.Type()
	if msgType == "" {
		return Call{}, sdkErrors.Wrap(ErrWrongMsgType, "empty")
	}

	// check other inputs
	if err := id.Valid(); err != nil {
		return Call{}, sdkErrors.Wrap(ErrWrongCallId, err.Error())
	}

	if uniqueID == "" {
		return Call{}, sdkErrors.Wrap(ErrWrongCallUniqueId, "empty")
	}

	if creatorAddr.Empty() {
		return Call{}, sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "creator: empty")
	}

	return Call{
		ID:       id,
		Creator:  creatorAddr,
		UniqueID: uniqueID,
		Approved: false,
		Executed: false,
		Rejected: false,
		Failed:   false,
		Msg:      msg,
		Error:    "",
		Height:   blockHeight,
		MsgRoute: msg.Route(),
		MsgType:  msg.Type(),
	}, nil
}
