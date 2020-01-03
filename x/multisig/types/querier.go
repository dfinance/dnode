package types

import (
	"fmt"
)

type UniqueReq struct {
	UniqueId string `json:"unique_id"`
}

type CallReq struct {
	CallId uint64 `json:"call_id"`
}

// Last id query response.
type LastIdRes struct {
	LastId uint64 `json:"last_id"`
}

// Last id response to string.
func (q LastIdRes) String() string {
	return fmt.Sprintf("Last id: %d", q.LastId)
}

// Get call query response.
type CallResp struct {
	Call  Call  `json:"call"`
	Votes Votes `json:"votes"`
}

// CallResp to string.
func (c CallResp) String() string {
	var votes string

	for i, v := range c.Votes {
		votes += v.String()

		if i != len(c.Votes)-1 {
			votes += ","
		}
	}

	return fmt.Sprintf("%sVotes: %s\n", c.Call.String(), votes)
}

type CallsResp []CallResp

func (calls CallsResp) String() string {
	var strCalls string

	for _, call := range calls {
		strCalls += call.String()
	}

	return strCalls
}
