package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName        = "multisig"
	RouterKey         = ModuleName
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
)

var (
	KeyDelimiter = []byte(":")
	QueuePrefix  = []byte("queue")
	// Key for storing last call ID
	LastCallIdKey = []byte("lastCallId")
)

// GetCallKey returns key for storing call objects.
func GetCallKey(callID dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("call"),
			[]byte(callID.String()),
		},
		KeyDelimiter,
	)
}

// GetUniqueIDKey returns key for storing callID by call's uniqueID.
func GetUniqueIDKey(callUniqueID string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("uniqueId"),
			[]byte(callUniqueID),
		},
		KeyDelimiter,
	)
}

// GetVotesKey returns key for storing call vote objects.
func GetVotesKey(callID dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("callVotes"),
			[]byte(callID.String()),
		},
		KeyDelimiter,
	)
}

// GetQueueKey return key for storing queue of callIDs.
func GetQueueKey(callID dnTypes.ID, blockHeight int64) []byte {
	return bytes.Join(
		[][]byte{
			QueuePrefix,
			sdk.Uint64ToBigEndian(uint64(blockHeight)),
			[]byte(callID.String()),
		},
		KeyDelimiter,
	)
}

// GetPrefixQueueKey returns queue of callIDs prefix key (used for iteration).
func GetPrefixQueueKey(blockHeight int64) []byte {
	return bytes.Join(
		[][]byte{
			QueuePrefix,
			sdk.Uint64ToBigEndian(uint64(blockHeight)),
		}, KeyDelimiter,
	)
}
