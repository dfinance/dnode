package types

import (
	"fmt"
)

// Default constants
const (
	ModuleName 		 = "multisig"

	DefaultRoute     = ModuleName
	DefaultCodespace = ModuleName
)

// Storage keys
var (
	LastCallId   	    = "lastCallId"
	LastExCallId    	= "lastExCallId"
	ExecutedCallByIdKey = "executedCall"
)

// Get a key to store call by id
func GetCallByIdKey(id uint64) []byte {
	return []byte(fmt.Sprintf("call:%d", id))
}

// Get a key to store executed call
func GetExCallByIdKey(id uint64) []byte {
	return []byte(fmt.Sprintf("ex_call:%d", id))
}

// Get a key to store votes for call by id
func GetKeyVotesById(id uint64) []byte {
	return []byte(fmt.Sprintf("votes:%d", id))
}