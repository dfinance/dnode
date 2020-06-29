package types

type AssetCode string

// Validate validates asset code.
func (a AssetCode) Validate() error {
	return nil
}

// String returns string enum representation.
func (a AssetCode) String() string {
	return string(a)
}
