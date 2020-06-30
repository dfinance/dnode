package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Withdraw is an info about reducing currency balance for the spender.
// swagger:model
type Withdraw struct {
	// Withdraw unique ID
	ID dnTypes.ID `json:"id" swaggertype:"string" example:"0"`
	// Target currency denom
	Denom string `json:"denom" example:"dfi"`
	// Amount of coins spender balance is decreased to
	Amount types.Int `json:"amount" swaggertype:"string" example:"100"`
	// Target account for reducing coin balance
	Spender types.AccAddress `json:"spender" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: spender account
	PegZoneSpender string `json:"pegzone_spender" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: ID
	PegZoneChainID string `json:"pegzone_chain_id" example:"testnet"`
	// Tx UNIX time [s]
	Timestamp int64 `json:"timestamp" format:"seconds" example:"1585295757"`
	// Tx hash
	TxHash string `json:"tx_hash" example:"fd82ce32835dfd7042808eaf6ff09cece952b9da20460fa462420a93607fa96f"`
}

func (withdraw Withdraw) String() string {
	return fmt.Sprintf("Withdraw:\n"+
		"  ID:             %s\n"+
		"  Denom:          %s\n"+
		"  Amount:         %s\n"+
		"  Spender:        %s\n"+
		"  PegZoneSpender: %s\n"+
		"  PegZoneChainID: %s\n"+
		"  Timestamp:      %d\n"+
		"  TxHash:         %s",
		withdraw.ID,
		withdraw.Denom,
		withdraw.Amount,
		withdraw.Spender,
		withdraw.PegZoneSpender,
		withdraw.PegZoneChainID,
		withdraw.Timestamp,
		withdraw.TxHash,
	)
}

// NewWithdraw creates a new Withdraw object.
func NewWithdraw(id dnTypes.ID, denom string, amount types.Int, spender types.AccAddress, pzSpender, pzChainID string, timestamp int64, txBytes []byte) Withdraw {
	hash := sha256.Sum256(txBytes)
	return Withdraw{
		ID:             id,
		Denom:          denom,
		Amount:         amount,
		Spender:        spender,
		PegZoneSpender: pzSpender,
		PegZoneChainID: pzChainID,
		Timestamp:      timestamp,
		TxHash:         hex.EncodeToString(hash[:]),
	}
}

// Withdraw slice.
type Withdraws []Withdraw

func (list Withdraws) String() string {
	var s strings.Builder
	for _, d := range list {
		s.WriteString(d.String() + "\n")
	}

	return s.String()
}
