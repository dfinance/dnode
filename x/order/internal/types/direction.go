package types

// Enum type to define bid/ask order type.
type Direction string

const (
	Bid Direction = "bid"
	Ask Direction = "ask"
)

// IsValid validates enum.
func (d Direction) IsValid() bool {
	if d == Bid || d == Ask {
		return true
	}

	return false
}

// String returns string enum representation.
func (d Direction) String() string {
	return string(d)
}
