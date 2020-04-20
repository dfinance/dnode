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

func GetProposalQueueKey(id uint64) []byte {
	return bytes.Join(
		[][]byte{
			ProposalQueuePrefix,
			sdk.Uint64ToBigEndian(id),
		},
		KeyDelimiter,
	)
}

func SplitProposalQueueKey(key []byte) (id uint64) {
	idBytes := key[len(ProposalQueuePrefix)+len(KeyDelimiter):]
	return binary.BigEndian.Uint64(idBytes)
}
