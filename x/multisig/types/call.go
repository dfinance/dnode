package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Call that will be executed itself, contains msg instances, that executing via router and hadler
type Call struct {
	// Creator
	Creator sdk.AccAddress   `json:"creator"`

	// Unique ID
	UniqueID string  		 `json:"unique_id"`

	// When call approved to execute
	Approved bool			 `json:"approved"`

	// Execution failed or executed
	Executed bool		     `json:"executed"`
	Failed   bool			 `json:"failed"`

	// If call was rejected
	Rejected bool			 `json:"rejected"`
	Error    string			 `json:"error"`

	// Msg to execute
	Msg MsMsg				 `json:"msg_data"`

	// Msg route
	MsgRoute string 		 `json:"msg_route"`

	// Msg type
	MsgType  string		     `json:"msg_type"`

	// Height when call submitted
	Height int64			 `json:"height"`
}

// Create new call instance
func NewCall(id uint64, uniqueID string, msg MsMsg, height int64, creator sdk.AccAddress) (Call, sdk.Error) {
	msgRoute := msg.Route()

	if msgRoute == "" {
		return Call{}, ErrEmptyRoute(id)
	}

	msgType  := msg.Type()

	if msgType == "" {
		return Call{}, ErrEmptyType(id)
	}

	return Call{
		Creator:  creator,
		UniqueID: uniqueID,
		Approved: false,
		Executed: false,
		Rejected: false,
		Failed:   false,
		Msg: 	  msg,
		Error:    "",
		Height:   height,
		MsgRoute: msg.Route(),
		MsgType:  msg.Type(),
	}, nil
}
