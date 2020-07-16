package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	EventTypeSubmitCall  = ModuleName + ".submit_call"
	EventTypeRemoveCall  = ModuleName + ".remove_call"
	EventTypeUpdateCall  = ModuleName + ".update_call"
	EventTypeConfirmVote = ModuleName + ".confirm_vote"
	EventTypeRevokeVote  = ModuleName + ".revoke_vote"
	//
	AttributeMsgType   = "msg_type"
	AttributeMsgRoute  = "msg_route"
	AttributeCallId    = "call_id"
	AttributeUniqueId  = "unique_id"
	AttributeSender    = "sender"
	AttributeCallState = "call_state"
	//
	AttributeValueApproved = "approved"
	AttributeValueRejected = "rejected"
	AttributeValueFailed   = "failed"
	AttributeValueExecuted = "executed"
)

// NewCallSubmittedEvent creates an Event on call submit (creation).
func NewCallSubmittedEvent(call Call) sdk.Event {
	return sdk.NewEvent(
		EventTypeSubmitCall,
		sdk.Attribute{Key: AttributeMsgType, Value: call.MsgType},
		sdk.Attribute{Key: AttributeMsgRoute, Value: call.MsgRoute},
		sdk.Attribute{Key: AttributeCallId, Value: call.ID.String()},
		sdk.Attribute{Key: AttributeUniqueId, Value: call.UniqueID},
		sdk.Attribute{Key: AttributeSender, Value: call.Creator.String()},
	)
}

// NewCallRemovedEvent creates an Event on call removal from the queue (deletion).
func NewCallRemovedEvent(callID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypeRemoveCall,
		sdk.Attribute{Key: AttributeCallId, Value: callID.String()},
	)
}

// NewCallStateChangedEvent creates an Event on call state change (approved, rejected, failed, executed).
func NewCallStateChangedEvent(callID dnTypes.ID, state string) sdk.Event {
	return sdk.NewEvent(
		EventTypeUpdateCall,
		sdk.Attribute{Key: AttributeCallId, Value: callID.String()},
		sdk.Attribute{Key: AttributeCallState, Value: state},
	)
}

// NewConfirmVoteEvent creates an Event on a new call vote (confirmed by sender).
func NewConfirmVoteEvent(callID dnTypes.ID, sender sdk.AccAddress) sdk.Event {
	return sdk.NewEvent(
		EventTypeConfirmVote,
		sdk.Attribute{Key: AttributeCallId, Value: callID.String()},
		sdk.Attribute{Key: AttributeSender, Value: sender.String()},
	)
}

// NewConfirmVoteEvent creates an Event on a new call vote (revoked by sender).
func NewRevokeVoteEvent(callID dnTypes.ID, sender sdk.AccAddress) sdk.Event {
	return sdk.NewEvent(
		EventTypeRevokeVote,
		sdk.Attribute{Key: AttributeCallId, Value: callID.String()},
		sdk.Attribute{Key: AttributeSender, Value: sender.String()},
	)
}
