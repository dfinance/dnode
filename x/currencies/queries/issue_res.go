package queries

import (
    "wings-blockchain/x/currencies/types"
    "fmt"
)

// Get currency query response
type QueryIssueRes struct {
    Issue types.Issue `json:"issue"`
}

func (q QueryIssueRes) String() string {
    return fmt.Sprintf("%s", q.Issue.String())
}
