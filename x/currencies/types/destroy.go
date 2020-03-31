// Implements destroy type for currencies module.
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"
)

// swagger:model
type Destroy struct {
	ID        types.Int        `json:"id" swaggertype:"string" example:"0"` // CallID
	ChainID   string           `json:"chainID" example:"testnet"`
	Symbol    string           `json:"symbol" example:"dfi"`
	Amount    types.Int        `json:"amount" swaggertype:"string" example:"100"`
	Spender   types.AccAddress `json:"spender" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	Recipient string           `json:"recipient" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	Timestamp int64            `json:"timestamp" format:"seconds" example:"1585295757"` // UNIX time
	TxHash    string           `json:"tx_hash" example:"fd82ce32835dfd7042808eaf6ff09cece952b9da20460fa462420a93607fa96f"`
}

func NewDestroy(id types.Int, chainID string, symbol string, amount types.Int, spender types.AccAddress, recipient string, txBytes []byte, timestamp int64) Destroy {
	hash := sha256.Sum256(txBytes)

	return Destroy{
		ID:        id,
		ChainID:   chainID,
		Symbol:    symbol,
		Amount:    amount,
		Spender:   spender,
		Recipient: recipient,
		Timestamp: timestamp,
		TxHash:    hex.EncodeToString(hash[:]),
	}
}

func (destroy Destroy) String() string {
	return fmt.Sprintf("Destroy: \n"+
		"\tChainID:   %s\n"+
		"\tID:        %s\n"+
		"\tSymbol:    %s\n"+
		"\tAmount:    %s\n"+
		"\tRecipient: %s\n"+
		"\tSpender:   %s\n"+
		"\tTxHash:    %s\n"+
		"\tTimestamp: %d\n",
		destroy.ChainID, destroy.ID,
		destroy.Symbol, destroy.Amount,
		destroy.Spender, destroy.Recipient,
		destroy.TxHash, destroy.Timestamp)
}

type Destroys []Destroy

func (destroys Destroys) String() string {
	var s strings.Builder
	for _, i := range destroys {
		s.WriteString(i.String())
	}

	return s.String()
}
