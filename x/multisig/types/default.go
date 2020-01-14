//
package types

import (
	"bytes"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default constants.
const (
	ModuleName = "multisig"

	RouterKey        = ModuleName
	DefaultCodespace = ModuleName

	IntervalToExecute int64 = 86400 // interval in blocks to execute proposal.
)

// Storage keys.
var (
	KeyDelimiter = []byte(":")

	LastCallId   = []byte("lastCallId")
	LastExCallId = []byte("lastExCallId")
	PrefixQueue  = []byte("callsQueue")
)

// Get a key to store call by id.
func GetCallByIdKey(id uint64) []byte {
	return []byte(fmt.Sprintf("call:%d", id))
}

// Get a key to store unique id.
func GetUniqueID(uniqueID string) []byte {
	return []byte(fmt.Sprintf("unique_id:%s", uniqueID))
}

// Get a key to store votes for call by id.
func GetKeyVotesById(id uint64) []byte {
	return []byte(fmt.Sprintf("votes:%d", id))
}

// Get a queue key for store call.
func GetQueueKey(id uint64, height int64) []byte {
	return bytes.Join(
		[][]byte{
			PrefixQueue,
			sdk.Uint64ToBigEndian(uint64(height)),
			sdk.Uint64ToBigEndian(id),
		},
		KeyDelimiter,
	)
}

// Get queue prefix based on height.
func GetPrefixQueue(height int64) []byte {
	return bytes.Join([][]byte{
		PrefixQueue,
		sdk.Uint64ToBigEndian(uint64(height)),
	}, KeyDelimiter)
}
