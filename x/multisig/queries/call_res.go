package queries

import (
	"wings-blockchain/x/multisig/types"
	"fmt"
)

// Get call query response
type QueryCallResp struct {
	Call  types.Call  `json:"call"`
	Votes types.Votes `json:"votes"`
}

func (q QueryCallResp) String() string {
	return fmt.Sprintf("Call:\n" +
		"\tCreator:   %s\n" +
	    "\tUnique ID: %s\n" +
		"\tApproved:  %t\n" +
		"\tRejected:  %t\n" +
		"\tError:     %s\n" +
		"\tHeight:    %d\n" +
		"\tMsg Route: %s\n" +
		"\tMsg Type:  %s\n" +
		"\tVotes:     %v\n",
		q.Call.Creator,  q.Call.UniqueID, q.Call.Approved,
		q.Call.Rejected, q.Call.Error,    q.Call.Height,
	    q.Call.MsgRoute, q.Call.MsgType,  q.Votes)
}

type QueryCallsResp []QueryCallResp

func (q QueryCallsResp) String() string {
    var s string
    for _, i := range q {
        s += i.String()
    }

    return s
}
