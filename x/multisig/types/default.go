package types

import (
	"encoding/binary"
)

const (
	ModuleName = "multisig"

	DefaultRoute = ModuleName
	DefaultCodespace = ModuleName
)

var (
	LastCallId = []byte("lastCallId")
	CallByIdKey = []byte("callByIdKey")
	VotesByIdKey = []byte("votesByIdKey")
)

func GetCallByIdKey(id uint64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, id)

	return append(MsgByIdKey, bs...)
}

func GetKeyVotesById(id uint64) []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, id)

	return append(VotesByIdKey, bs...)
}