package queries

import (
	"github.com/WingsDao/wings-blockchain/x/currencies/types"
)

// Get currency query response
type QueryIssueRes struct {
	Issue types.Issue `json:"issue"`
}

func (q QueryIssueRes) String() string {
	return q.Issue.String()
}
