package queries

import (
	"strings"

	"wings-blockchain/x/currencies/types"
)

// Get currency query response
type QueryDestroyRes struct {
	Destroy types.Destroy `json:"destroy"`
}

func (q QueryDestroyRes) String() string {
	return q.Destroy.String()
}

type QueryDestroysRes []QueryDestroyRes

func (q QueryDestroysRes) String() string {
	s := strings.Builder{}
	for _, i := range q {
		s.WriteString(i.String())
	}

	return s.String()
}
