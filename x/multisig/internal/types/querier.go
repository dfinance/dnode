package types

import (
	"fmt"
	"strings"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryCalls        = "calls"
	QueryCall         = "call"
	QueryCallByUnique = "callByUnique"
	QueryLastId       = "lastId"
)

// Client request for call by call ID.
type CallReq struct {
	CallID dnTypes.ID `json:"call_id" yaml:"call_id"`
}

// Client request for call by call uniqueID.
type CallByUniqueIdReq struct {
	UniqueID string `json:"unique_id" yaml:"unique_id"`
}

// Client response for last call ID.
type LastCallIdResp struct {
	LastID dnTypes.ID `json:"last_id" yaml:"last_id"`
}

func (r LastCallIdResp) String() string {
	return fmt.Sprintf("Last callID: %s", r.LastID.String())
}

// Client response for call with votes.
type CallResp struct {
	// Call info
	Call Call `json:"call" yaml:"call"`
	// Voted accounts addresses
	Votes Votes `json:"votes" yaml:"votes" swaggertype:"array,string"`
}

func (r CallResp) String() string {
	strBuilder := strings.Builder{}
	for i, v := range r.Votes {
		strBuilder.WriteString(v.String())
		if i < len(r.Votes)-1 {
			strBuilder.WriteString(", ")
		}
	}

	return fmt.Sprintf("%s\nVotes: [%s]", r.Call.String(), strBuilder.String())
}

// Client response for multiple calls with votes.
type CallsResp []CallResp

func (r CallsResp) String() string {
	strBuilder := strings.Builder{}
	for i, call := range r {
		strBuilder.WriteString(fmt.Sprintf("[%d] %s", i, call.String()))
		if i < len(r)-1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}
