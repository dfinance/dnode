package types

import (
	"fmt"
)

// Request to get call by unique id.
type UniqueReq struct {
	UniqueId string `json:"unique_id"`
}

// Request to get call by call id.
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

// Type to get a call as response with votes.
type CallResp struct {
	Call  Call  `json:"call"`  // Call info
	Votes Votes `json:"votes" swaggertype:"array,string"` // Accounts address array
}

// Call response to string.
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

// Slice of call responses (in case of multiplay calls to response).
type CallsResp []CallResp

// Call responses to string.
func (calls CallsResp) String() string {
	var strCalls string

	for _, call := range calls {
		strCalls += call.String()
	}

	return strCalls
}
