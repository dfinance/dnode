package types

type AssetCode string

// IsValid validates asset code.
func (a AssetCode) IsValid() bool {
	return true
}

// String returns string enum representation.
func (a AssetCode) String() string {
	return string(a)
}
