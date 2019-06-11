package queries

import "fmt"

// Last id query response
type QueryLastIdRes struct {
	LastId uint64 `json:"lastId"`
}

func (q QueryLastIdRes) String() string {
	return fmt.Sprintf("Last id: %d", q.LastId)
}
