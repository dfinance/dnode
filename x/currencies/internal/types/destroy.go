package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Destroy is an info about destroying currency for the spender.
// swagger:model
type Destroy struct {
	// Destroy unique ID
	ID dnTypes.ID `json:"id" swaggertype:"string" example:"0"`
	// Target currency denom
	Denom string `json:"denom" example:"dfi"`
	// Amount of coins spender balance is decreased to
	Amount types.Int `json:"amount" swaggertype:"string" example:"100"`
	// Target account for reducing coin balance
	Spender types.AccAddress `json:"spender" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: spender account
	Recipient string `json:"recipient" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: ID
	ChainID string `json:"chainID" example:"testnet"`
	// Tx UNIX time [s]
	Timestamp int64 `json:"timestamp" format:"seconds" example:"1585295757"`
	// Tx hash
	TxHash string `json:"tx_hash" example:"fd82ce32835dfd7042808eaf6ff09cece952b9da20460fa462420a93607fa96f"`
}

func (destroy Destroy) String() string {
	return fmt.Sprintf("Destroy:\n"+
		"  ID:        %s\n"+
		"  Denom:     %s\n"+
		"  Amount:    %s\n"+
		"  Spender:   %s\n"+
		"  Recipient: %s\n"+
		"  ChainID:   %s\n"+
		"  Timestamp: %d\n"+
		"  TxHash:    %s",
		destroy.ID,
		destroy.Denom,
		destroy.Amount,
		destroy.Spender,
		destroy.Recipient,
		destroy.ChainID,
		destroy.Timestamp,
		destroy.TxHash,
	)
}

// NewDestroy creates a new Destroy object.
func NewDestroy(id dnTypes.ID, denom string, amount types.Int, spender types.AccAddress, recipient, chainID string, timestamp int64, txBytes []byte) Destroy {
	hash := sha256.Sum256(txBytes)
	return Destroy{
		ID:        id,
		Denom:     denom,
		Amount:    amount,
		Spender:   spender,
		Recipient: recipient,
		ChainID:   chainID,
		Timestamp: timestamp,
		TxHash:    hex.EncodeToString(hash[:]),
	}
}

// Destroy slice.
type Destroys []Destroy

func (list Destroys) String() string {
	var s strings.Builder
	for _, d := range list {
		s.WriteString(d.String() + "\n")
	}

	return s.String()
}
