package queries

import "fmt"

type QueryLastIdRes struct {
	LastId uint64 `json:"lastId"`
}

func (q QueryLastIdRes) String() string {
	return fmt.Sprintf("%d", q.LastId)
}
