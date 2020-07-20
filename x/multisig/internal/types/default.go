package types

import (
	"bytes"
	"encoding/binary"
	"fmt"

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
	CallPrefix   = []byte("call")
	QueuePrefix  = []byte("queue")
	// Key for storing last call ID
	LastCallIdKey = []byte("lastCallId")
)

// GetCallKey returns key for storing call objects.
func GetCallKey(callID dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			CallPrefix,
			[]byte(callID.String()),
		},
		KeyDelimiter,
	)
}

// GetCallKeyPrefix returns key prefix for call objects iteration.
func GetCallKeyPrefix() []byte {
	return append(CallPrefix, KeyDelimiter...)
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

// MustParseQueueKey parses queue storage key.
func MustParseQueueKey(key []byte) (callID dnTypes.ID, blockHeight int64) {
	values := bytes.Split(key, KeyDelimiter)
	if len(values) != 3 {
		panic(fmt.Errorf("key %q: invalid splitted length %d", string(key), len(values)))
	}

	if !bytes.Equal(values[0], QueuePrefix) {
		panic(fmt.Errorf("key %q: value[0] %q: wrong prefix", string(key), string(values[0])))
	}

	blockHeight = int64(binary.BigEndian.Uint64(values[1]))

	id, err := dnTypes.NewIDFromString(string(values[2]))
	if err != nil {
		panic(fmt.Errorf("key %q: value[2] %q: %w", string(key), string(values[2]), err))
	}
	callID = id

	return
}
