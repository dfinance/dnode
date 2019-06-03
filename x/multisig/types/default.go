package types

import (
	"encoding/binary"
)

// Default constants
const (
	ModuleName 		 = "multisig"

	DefaultRoute     = ModuleName
	DefaultCodespace = ModuleName
)

// Storage keys
var (
	LastCallId   = []byte("lastCallId")
	CallByIdKey  = []byte("callByIdKey")
	VotesByIdKey = []byte("votesByIdKey")
)

// Get a key to store call by id
func GetCallByIdKey(id uint64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, id)

	return append(MsgByIdKey, bs...)
}

// Get a key to store votes for call by id
func GetKeyVotesById(id uint64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, id)

	return append(VotesByIdKey, bs...)
}