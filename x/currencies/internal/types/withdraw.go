package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Withdraw is an info about reducing currency balance for the spender.
// swagger:model
type Withdraw struct {
	// Withdraw unique ID
	ID dnTypes.ID `json:"id" yaml:"id" format:"string representation for big.Uint" swaggertype:"string" example:"0"`
	// Target currency coin
	Coin sdk.Coin `json:"coin" yaml:"coin" swaggertype:"string" example:"100dfi"`
	// Target account for reducing coin balance
	Spender sdk.AccAddress `json:"spender" yaml:"spender" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: spender account
	PegZoneSpender string `json:"pegzone_spender" yaml:"pegzone_spender" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	// Second blockchain: ID
	PegZoneChainID string `json:"pegzone_chain_id" yaml:"pegzone_chain_id" example:"testnet"`
	// Tx UNIX time [s]
	Timestamp int64 `json:"timestamp" yaml:"timestamp" format:"seconds" example:"1585295757"`
	// Tx hash
	TxHash string `json:"tx_hash" yaml:"tx_hash" example:"fd82ce32835dfd7042808eaf6ff09cece952b9da20460fa462420a93607fa96f"`
}

func (withdraw Withdraw) String() string {
	return fmt.Sprintf("Withdraw:\n"+
		"  ID:             %s\n"+
		"  Coin:           %s\n"+
		"  Payer:        %s\n"+
		"  PegZoneSpender: %s\n"+
		"  PegZoneChainID: %s\n"+
		"  Timestamp:      %d\n"+
		"  TxHash:         %s",
		withdraw.ID,
		withdraw.Coin.String(),
		withdraw.Spender,
		withdraw.PegZoneSpender,
		withdraw.PegZoneChainID,
		withdraw.Timestamp,
		withdraw.TxHash,
	)
}

// NewWithdraw creates a new Withdraw object.
func NewWithdraw(id dnTypes.ID, coin sdk.Coin, spender sdk.AccAddress, pzSpender, pzChainID string, timestamp int64, txBytes []byte) Withdraw {
	hash := sha256.Sum256(txBytes)
	return Withdraw{
		ID:             id,
		Coin:           coin,
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
	for i, d := range list {
		s.WriteString(d.String())
		if i < len(list)-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}
