package types

import (
	"bytes"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ProposalIDKey       = []byte("proposal_id")
	ProposalQueuePrefix = []byte("proposal_queue")
)

// GetProposalQueueKey returns gov proposal queue storage key.
func GetProposalQueueKey(id uint64) []byte {
	return bytes.Join(
		[][]byte{
			ProposalQueuePrefix,
			sdk.Uint64ToBigEndian(id),
		},
		KeyDelimiter,
	)
}

// SplitProposalQueueKey returns gov proposal queue ID for the key.
func SplitProposalQueueKey(key []byte) (id uint64) {
	idBytes := key[len(ProposalQueuePrefix)+len(KeyDelimiter):]
	return binary.BigEndian.Uint64(idBytes)
}
