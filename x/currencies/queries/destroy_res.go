package queries

import (
    "wings-blockchain/x/currencies/types"
    "fmt"
)

// Get currency query response
type QueryDestroyRes struct {
    Destroy types.Destroy `json:"destroy"`
}

func (q QueryDestroyRes) String() string {
    return fmt.Sprintf("%s", q.Destroy.String())
}

type QueryDestroysRes []QueryDestroyRes

func (q QueryDestroysRes) String() string {
    var s string
    for _, i := range q {
        s += i.String()
    }

    return s
}

