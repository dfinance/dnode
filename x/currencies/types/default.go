package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	ModuleName 	  				      	  = "currencies"

	DefaultRoute 	  					  = ModuleName
	DefaultCodespace  sdk.CodespaceType   = ModuleName
	DefaultParamspace 					  = ModuleName
)

var (
	ValidatorsCountKey = []byte("validators_count")
)
