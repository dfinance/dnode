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

// Equal check whether d and d2 are equal.
func (d Direction) Equal(d2 Direction) bool {
	return d.String() == d2.String()
}

// String returns string enum representation.
func (d Direction) String() string {
	return string(d)
}

// NewDirectionRaw creates a new Direction object without checks.
func NewDirectionRaw(str string) Direction {
	return Direction(str)
}
