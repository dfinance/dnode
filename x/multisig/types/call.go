// Create call message type.
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core/msmodule"
)

// Call that will be executed itself, contains msg instances, that executing via router and handler.
type Call struct {
	Creator  sdk.AccAddress `json:"creator" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	MsgID    uint64         `json:"msg_id" example:"0"`           // Call ID
	UniqueID string         `json:"unique_id" example:"issue1"`   // Call uniqueID
	Approved bool           `json:"approved"`                     // Call is approved to execute
	Executed bool           `json:"executed"`                     // Call was executed
	Failed   bool           `json:"failed"`                       // Call failed to execute
	Rejected bool           `json:"rejected"`                     // Call was rejected
	Error    string         `json:"error"`                        // Call failed reason
	Msg      msmodule.MsMsg `json:"msg_data"`                     // Message to execute
	MsgRoute string         `json:"msg_route" example:"oracle"`   // Message route
	MsgType  string         `json:"msg_type" example:"add_asset"` // Message type
	Height   int64          `json:"height" example:"1"`           // BlockHeight when call was submitted
}

// Create new call instance.
func NewCall(id uint64, uniqueID string, msg msmodule.MsMsg, height int64, creator sdk.AccAddress) (Call, error) {
	msgRoute := msg.Route()

	if msgRoute == "" {
		return Call{}, ErrEmptyRoute
	}

	msgType := msg.Type()

	if msgType == "" {
		return Call{}, sdkErrors.Wrapf(ErrEmptyType, "%d", id)
	}

	return Call{
		Creator:  creator,
		MsgID:    id,
		UniqueID: uniqueID,
		Approved: false,
		Executed: false,
		Rejected: false,
		Failed:   false,
		Msg:      msg,
		Error:    "",
		Height:   height,
		MsgRoute: msg.Route(),
		MsgType:  msg.Type(),
	}, nil
}

// Convert call to string representation.
func (c Call) String() string {
	return fmt.Sprintf("Call:\n"+
		"\tCreator:   %s\n"+
		"\tUnique ID: %s\n"+
		"\tApproved:  %t\n"+
		"\tRejected:  %t\n"+
		"\tError:     %s\n"+
		"\tHeight:    %d\n"+
		"\tMsg Route: %s\n"+
		"\tMsg Type:  %s\n",
		c.Creator, c.UniqueID, c.Approved,
		c.Rejected, c.Error, c.Height,
		c.MsgRoute, c.MsgType)
}
