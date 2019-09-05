package types

import (
    "github.com/cosmos/cosmos-sdk/types"
    "fmt"
    "crypto/sha256"
    "encoding/hex"
)

type Destroy struct {
    ID         types.Int        `json:"id"`
    CurrencyId types.Int        `json:"currencyId"`
    ChainID    string           `json:"chainID"`
    Symbol     string           `json:"symbol"`
    Amount     types.Int        `json:"amount"`
    Spender    types.AccAddress `json:"spender"`
    Recipient  string           `json:"recipient"`
    Timestamp  int64            `json:"timestamp"`
    TxHash     string           `json:"tx_hash"`
}

func NewDestroy(id types.Int, currencyId types.Int, chainID string, symbol string, amount types.Int, spender types.AccAddress, recipient string, txBytes []byte, timestamp int64) Destroy {
    hash := sha256.Sum256(txBytes)

    return Destroy{
        ID:         id,
        CurrencyId: currencyId,
        ChainID:    chainID,
        Symbol:     symbol,
        Amount:     amount,
        Spender:    spender,
        Recipient:  recipient,
        Timestamp:  timestamp,
        TxHash:     hex.EncodeToString(hash[:]),
    }
}

func (destroy Destroy) String() string {
    return fmt.Sprintf("Destroy: \n" +
        "\tChainID:   %s\n" +
        "\tID:        %s\n" +
        "\tSymbol:    %s\n" +
        "\tAmount:    %s\n" +
        "\tRecipient: %s\n" +
        "\tSpender:   %s\n" +
        "\tTxHash:    %s\n" +
        "\tTimestamp: %d\n",
        destroy.ChainID, destroy.ID,
        destroy.Symbol,  destroy.Amount,
        destroy.Spender, destroy.Recipient,
        destroy.TxHash,  destroy.Timestamp)
}

