package queries

import "fmt"

type QueryMinMaxRes struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

func (q QueryMinMaxRes) String() string {
	return fmt.Sprintf("Min: %d\n" +
		"Max: %d\n", q.Min, q.Max)
}
