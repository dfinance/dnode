package types

type Direction string

const (
	Bid Direction = "bid"
	Ask Direction = "ask"
)

func (d Direction) IsValid() bool {
	if d == Bid || d == Ask {
		return true
	}

	return false
}

func (d Direction) String() string {
	return string(d)
}
