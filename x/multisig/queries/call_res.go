package queries

import (
	"wings-blockchain/x/multisig/types"
	"fmt"
)

// Get call query response
type QueryCallResp struct {
	Call types.Call	`json:"call"`
}

func (q QueryCallResp) String() string {
	return fmt.Sprintf("Call:\n" +
		"\tCreator:   %s\n" +
		"\tApproved:  %t\n" +
		"\tRejected:  %t\n" +
		"\tError:     %s\n" +
		"\tHeight:    %d\n" +
		"\tMsg Route: %s\n" +
		"\tMsg Type:  %s\n",
		q.Call.Creator.String(), q.Call.Approved,
		q.Call.Rejected, q.Call.Error,
		q.Call.Height, q.Call.MsgRoute, q.Call.MsgType)
}