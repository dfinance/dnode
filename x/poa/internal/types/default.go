package types

const (
	ModuleName        = "poa"
	RouterKey         = ModuleName
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
)

var (
	// Key for storing validators counter
	ValidatorsCountKey = []byte("validatorsCount")
	// Key for storing validator objects
	ValidatorsListKey = []byte("validators")
)
